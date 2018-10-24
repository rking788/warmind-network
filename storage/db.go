package storage

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	raven "github.com/getsentry/raven-go"
	"github.com/kpango/glg"
	_ "github.com/lib/pq" // Only want to import the interface here
	"github.com/rking788/warmind-network/models"
)

const (
	// UnknownClassTable is the name of the table that will hold all the unknown class values provided by Alexa
	UnknownClassTable = "unknown_classes"
	// UnknownItemTable is the name of the table that will hold the unknown item name values passed by Alexa
	UnknownItemTable = "unknown_items"
)

// LookupDB is a wrapper around the database connection pool that stores the commonly used queries
// as prepared statements.
type LookupDB struct {
	Database                 *sql.DB
	HashFromNameStmt         *sql.Stmt
	NameFromHashStmt         *sql.Stmt
	EngramHashStmt           *sql.Stmt
	ItemMetadataStmt         *sql.Stmt
	RandomJokeStmt           *sql.Stmt
	InsertLoadoutStmt        *sql.Stmt
	UpdateLoadoutStmt        *sql.Stmt
	SelectLoadoutStmt        *sql.Stmt
	SelectLoadoutByNameStmt  *sql.Stmt
	SelectActivityByHashStmt *sql.Stmt
}

var db1 *LookupDB
var dbURL string

// InitDatabase is in charge of preparing any Statements that will be commonly used as well
// as setting up the database connection pool.
func InitDatabase() error {

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("DB errror: %s", err.Error())
		return err
	}

	stmt, err := db.Prepare("SELECT item_hash FROM items WHERE item_name = $1 AND item_type_name NOT IN ('Material Exchange', '') ORDER BY max_stack_size DESC LIMIT 1")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("DB error: %s", err.Error())
		return err
	}
	nameFromHashStmt, err := db.Prepare("SELECT item_name FROM items WHERE item_hash = $1 LIMIT 1")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("DB prepare error: %s", err.Error())
		return err
	}

	// 8 is the item_type value for engrams
	engramHashStmt, err := db.Prepare("SELECT item_hash FROM items WHERE item_name LIKE '%engram%'")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("DB prepare error: %s", err.Error())
		return err
	}

	itemMetadataStmt, err := db.Prepare("SELECT item_hash, item_name, tier_type, class_type, bucket_type_hash FROM items")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("DB error: %s", err.Error())
		return err
	}

	randomJokeStmt, err := db.Prepare("SELECT * FROM jokes offset random() * (SELECT COUNT(*) FROM jokes) LIMIT 1")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("DB prepare error: %s", err.Error())
		return err
	}

	insertLoadoutStmt, err := db.Prepare("INSERT INTO loadouts VALUES ($1,$2,$3)")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error preparing insert loadout statement: %s", err.Error())
		return err
	}

	updateLoadoutStmt, err := db.Prepare("UPDATE loadouts SET loadout=$1 WHERE bungie_membership_id=$2 AND name=$3")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed preparing update loadout statement: %s", err.Error())
		return err
	}

	selectLoadoutStmt, err := db.Prepare("SELECT name FROM loadouts WHERE bungie_membership_id=$1")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error preparing the select loadout statement: %s", err.Error())
		return err
	}

	selectLoadoutByNameStmt, err := db.Prepare("SELECT loadout FROM loadouts WHERE bungie_membership_id=$1 AND name=$2")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error preparing the select loadout by name statement: %s", err.Error())
		return err
	}

	selectActivityByHashStmt, err := db.Prepare("SELECT a.hash, a.name, a.description, a.light_level, a.is_playlist, a.is_pvp, d.hash, d.name, d.description, p.hash, p.name, p.description, at.hash, at.name, at.description, am.hash, am.name, am.mode_type, am.category, am.tier, am.is_team_based, a.rewards, a.challenges, a.modifiers FROM activities a, destinations d, places p, activity_types at, activity_modes am WHERE a.hash = $1 AND d.hash = a.destination_hash AND p.hash = a.place_hash AND a.activity_type_hash = at.hash AND a.direct_activity_mode_hash = am.hash")
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error prepraing the select activity by hash statement: %s", err.Error())
		return err
	}

	db1 = &LookupDB{
		Database:                 db,
		HashFromNameStmt:         stmt,
		NameFromHashStmt:         nameFromHashStmt,
		EngramHashStmt:           engramHashStmt,
		ItemMetadataStmt:         itemMetadataStmt,
		RandomJokeStmt:           randomJokeStmt,
		InsertLoadoutStmt:        insertLoadoutStmt,
		UpdateLoadoutStmt:        updateLoadoutStmt,
		SelectLoadoutStmt:        selectLoadoutStmt,
		SelectLoadoutByNameStmt:  selectLoadoutByNameStmt,
		SelectActivityByHashStmt: selectActivityByHashStmt,
	}

	return nil
}

// GetDBConnection is a helper for getting a connection to the DB based on
// environment variables or some other method.
func GetDBConnection() (*LookupDB, error) {

	if db1 == nil {
		glg.Info("Initializing db!")
		err := InitDatabase()
		if err != nil {
			raven.CaptureError(err, nil)
			glg.Errorf("Failed to initialize the database: %s", err.Error())
			return nil, err
		}
	}

	return db1, nil
}

// FindEngramHashes is responsible for querying all of the item_hash values that represent engrams
// and returning them in a map for quick lookup later.
func FindEngramHashes() (map[uint]bool, error) {

	result := make(map[uint]bool)

	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}

	rows, err := db.EngramHashStmt.Query()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var hash uint
		rows.Scan(&hash)
		result[hash] = true
	}

	return result, nil
}

// LoadItemMetadata will load all rows from the database for all items loaded out of the manifest.
// Only the required columns will be loaded into memory that need to be used later for common operations.
func LoadItemMetadata() (map[uint]*models.ItemMetadata, error) {

	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}

	rows, err := db.ItemMetadataStmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itemMetadata := make(map[uint]*models.ItemMetadata)
	for rows.Next() {
		var hash uint
		itemMeta := models.ItemMetadata{}
		rows.Scan(&hash, &itemMeta.Name, &itemMeta.TierType, &itemMeta.ClassType, &itemMeta.BucketHash)

		itemMetadata[hash] = &itemMeta
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return itemMetadata, nil
}

// GetItemNameFromHash is in charge of querying the database and reading
// the item name value for the given item hash.
func GetItemNameFromHash(itemHash string) (string, error) {

	db, err := GetDBConnection()
	if err != nil {
		return "", err
	}

	row := db.NameFromHashStmt.QueryRow(itemHash)

	var name string
	err = row.Scan(&name)

	if err == sql.ErrNoRows {
		glg.Warnf("Didn't find any transferrable items with that hash: %s", itemHash)
		return "", errors.New("No items found")
	} else if err != nil {
		return "", errors.New(err.Error())
	}

	return name, nil
}

// SaveLoadout is responsible for persisting the provided serialized loadout to the database.
func SaveLoadout(loadout models.Loadout, membershipID, name string) error {

	db, err := GetDBConnection()
	if err != nil {
		return err
	}

	persistedLoadout := loadout.ToPersistedLoadout()
	persistedBytes, err := json.Marshal(persistedLoadout)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to marshal the loadout to JSON: %s", err.Error())
		return err
	}

	_, err = db.InsertLoadoutStmt.Exec(membershipID, name, string(persistedBytes))

	return err
}

// UpdateLoadout can be used to update an existing loadout by membership ID and loadout name
// this should be used after confirming with the user that they want to update a loadout
// with a spepcific name.
func UpdateLoadout(loadout models.Loadout, membershipID, name string) error {

	db, err := GetDBConnection()
	if err != nil {
		return err
	}

	persistedLoadout := loadout.ToPersistedLoadout()
	persistedBytes, err := json.Marshal(persistedLoadout)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to marshal the loadout to JSON: %s", err.Error())
		return err
	}

	_, err = db.UpdateLoadoutStmt.Exec(string(persistedBytes), membershipID, name)

	return err
}

// SelectLoadout is responsible for querying the database for a loadout with the
// provided membership ID and loadout name. The return value is the JSON string for
// the loadout requested.
func SelectLoadout(membershipID, name string) (models.PersistedLoadout, error) {

	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}

	row := db.SelectLoadoutByNameStmt.QueryRow(membershipID, name)

	var loadoutJSON string
	err = row.Scan(&loadoutJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var persistedLoadout models.PersistedLoadout
	err = json.NewDecoder(bytes.NewReader([]byte(loadoutJSON))).Decode(&persistedLoadout)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to decode JSON: %s", err.Error())
		return nil, err
	}

	return persistedLoadout, nil
}

// SelectLoadouts is responsible for querying the database for a loadout with the
// provided membership ID and loadout name. The return value is a slice of the names
// of all loadouts saved for the provided user.
func SelectLoadouts(membershipID string) ([]string, error) {

	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}

	rows, err := db.SelectLoadoutStmt.Query(membershipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0, 10)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			raven.CaptureError(err, nil)
			glg.Errorf("Error retrieving loadout name from DB...%s\n", err.Error())
			continue
		}
		result = append(result, name)
	}

	return result, nil
}

// InsertUnknownValueIntoTable is a helper method for inserting a value into the specified table.
// This is used when a value for a slot type is not usable. For example when a class name for
// a character is not a valid Destiny class name.
func InsertUnknownValueIntoTable(value, tableName string) {

	conn, err := GetDBConnection()
	if err != nil {
		return
	}

	conn.Database.Exec("INSERT INTO "+tableName+" (value) VALUES(?)", value)
}

// GetRandomJoke will return a setup, punchline, and possibly an error for a random Destiny related joke.
func GetRandomJoke() (string, string, error) {

	db, err := GetDBConnection()
	if err != nil {
		return "", "", err
	}

	row := db.RandomJokeStmt.QueryRow()

	var setup string
	var punchline string
	err = row.Scan(&setup, &punchline)

	return setup, punchline, nil
}

type modifiersTemp struct {
	Hash uint `json:"activityModifierHash"`
}

// GetActivity will load an activity and all required associated properties from other tables such as
// the places, destinations, and modifiers tables.
func GetActivity(hash uint) (*models.Activity, error) {
	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}
	hash = 2752743635

	// SELECT a.hash, a.name, a.description, a.light_level, a.is_playlist, a.is_pvp, d.hash, d.name, d.description,
	// 		p.hash, p.name, p.description, at.hash, at.name, at.description, am.hash, am.name, am.mode_type,
	// 		am.category, am.tier, am.is_team_based, a.rewards, a.challenges, a.modifiers
	// FROM activities a, destinations d, places p, activity_types at, activity_modes am
	// WHERE a.hash = 2752743635
	// 	AND d.hash = a.destination_hash
	// 	AND p.hash = a.place_hash
	// 	AND a.activity_type_hash = at.hash
	// 	AND a.direct_activity_mode_hash = am.hash;

	activity := models.Activity{}

	activity.Destination = &models.Destination{}
	activity.Place = &models.Place{}
	activity.ActivityMode = &models.ActivityMode{}
	activity.ActivityType = &models.ActivityType{}

	row := db.SelectActivityByHashStmt.QueryRow(hash)

	if row == nil {
		fmt.Println("Found nil row from query")
	}

	var rewardsJSON string
	var challengesJSON string
	var modifiersJSON string

	err = row.Scan(&activity.Hash, &activity.Name, &activity.Description,
		&activity.LightLevel, &activity.IsPlaylist, &activity.IsPVP,
		&activity.Destination.Hash, &activity.Destination.Name,
		&activity.Destination.Description, &activity.Place.Hash, &activity.Place.Name,
		&activity.Place.Description, &activity.ActivityType.Hash, &activity.ActivityType.Name,
		&activity.ActivityType.Description, &activity.ActivityMode.Hash, &activity.ActivityMode.Name,
		&activity.ActivityMode.ModeType, &activity.ActivityMode.Category,
		&activity.ActivityMode.Tier, &activity.ActivityMode.IsTeamBased, &rewardsJSON, &challengesJSON, &modifiersJSON)
	if err != nil {
		fmt.Printf("Error scanning activity results: %v\n", err.Error())
	}

	fmt.Printf("Rewards: %v\nChallenges: %v\nModifiers: %v\n\n", rewardsJSON, challengesJSON, modifiersJSON)
	err = json.Unmarshal([]byte(rewardsJSON), &activity.Rewards)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(challengesJSON), &activity.Challenges)
	if err != nil {
		return nil, err
	}

	tempModifiers := make([]*modifiersTemp, 0, 10)
	err = json.Unmarshal([]byte(modifiersJSON), &tempModifiers)
	activity.Modifiers = make([]uint, 0, len(tempModifiers))
	for _, modifier := range tempModifiers {
		activity.Modifiers = append(activity.Modifiers, modifier.Hash)
	}

	if err == sql.ErrNoRows {
		glg.Warnf("Didn't find any activities with that hash: %s", hash)
		return nil, errors.New("No activities found")
	} else if err != nil {
		return nil, errors.New(err.Error())
	}

	return &activity, nil
}
