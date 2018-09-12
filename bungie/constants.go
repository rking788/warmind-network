package bungie

// Constant API endpoints
const (
	//GetCurrentAccountEndpoint = "http://localhost:8000/account.json"
	//ItemsEndpointFormat       = "http://localhost:8000/%d/%s/items.json"
	//TransferItemEndpointURL           = "http://localhost:8000/d1/Platform/Destiny/TransferItem/"
	//EquipItemEndpointURL              = "http://localhost:8000/d1/Platform/Destiny/EquipItem/"
	GetCurrentAccountEndpoint            = "https://www.bungie.net/Platform/User/GetCurrentBungieAccount/"
	GetMembershipsForCurrentUserEndpoint = "https://www.bungie.net/Platform/User/GetMembershipsForCurrentUser/"
	GetProfileEndpointFormat             = "https://www.bungie.net/platform/Destiny2/%d/Profile/%s/"
	TransferItemEndpointURL              = "https://www.bungie.net/Platform/Destiny2/Actions/Items/TransferItem/"
	EquipSingleItemEndpointURL           = "https://www.bungie.net/Platform/Destiny2/Actions/Items/EquipItem/"
	EquipMultiItemsEndpointURL           = "https://www.bungie.net/Platform/Destiny2/Actions/Items/EquipItems/"
)

// Random constant game data
const (
	MaxItemsPerBucket = 10
)

// Component constant values that are needed for certain Bungie API requests that specify which
// collections of values should be returned in the response.
const (
	ProfilesComponent              = "100"
	VendorReceiptsComponent        = "101"
	ProfileInventoriesComponent    = "102"
	ProfileCurrenciesComponent     = "103"
	CharactersComponent            = "200"
	CharacterInventoriesComponent  = "201"
	CharacterProgressionsComponent = "202"
	CharacterRenderDataComponent   = "203"
	CharacterActivitiesComponent   = "204"
	CharacterEquipmentComponent    = "205"
	ItemInstancesComponent         = "300"
	ItemObjectivesComponent        = "301"
	ItemPerksComponent             = "302"
	ItemRenderDataComponent        = "303"
	ItemStatsComponent             = "304"
	ItemSocketsComponent           = "305"
	ItemTalentGridsComponent       = "306"
	ItemCommonDataComponent        = "307"
	ItemPlugStatesComponent        = "308"
	VendorsComponent               = "400"
	VendorCategoriesComponent      = "401"
	VendorSalesComponent           = "402"
	KiosksComponent                = "500"
)

// Destiny.TierType
const (
	UnknownTier  = 0
	CurrencyTier = 1
	BasicTier    = 2
	CommonTier   = 3
	RareTier     = 4
	SuperiorTier = 5
	ExoticTier   = 6
)

// Destiny.TansferStatuses
const (
	CanTransfer         = 0
	ItemIsEquipped      = 1
	NotTransferrable    = 2
	NoRoomInDestination = 4
)

// Destiny.ItemLocation
const (
	UnknownLoc    = 0
	InventoryLoc  = 1
	VaultLoc      = 2
	VendorLoc     = 3
	PostmasterLoc = 4
)

// Hash values for different class types 'classHash' JSON key
const (
	WARLOCK = 2271682572
	TITAN   = 3655393761
	HUNTER  = 671679327
)

var classHashToName = map[uint]string{
	WARLOCK: "warlock",
	TITAN:   "titan",
	HUNTER:  "hunter",
}

var classNameToHash = map[string]uint{
	"warlock": WARLOCK,
	"titan":   TITAN,
	"hunter":  HUNTER,
}

// SubclassHash holds the hash representing the exact subclass "item". These are represented with
// items and a subclass bucket definition.
type SubclassHash uint

// Subclass hashes (from the subclass inventory bucket)
const (
	Striker    SubclassHash = 2958378809
	Sunbreaker SubclassHash = 3105935002
	Sentinel   SubclassHash = 3382391785

	Arcstrider   SubclassHash = 1334959255
	Gunslinger   SubclassHash = 3635991036
	Nightstalker SubclassHash = 3225959819

	Stormcaller SubclassHash = 1751782730
	Dawnblade   SubclassHash = 3481861797
	Voidwalker  SubclassHash = 3887892656
)

// Class Enum value passed in some of the Destiny API responses
const (
	TitanClassType   = 0
	HunterClassType  = 1
	WarlockEnum      = 2
	UnknownClassType = 3
)

// Hash values for Race types 'raceHash' JSON key
const (
	AWOKEN = 2803282938
	HUMAN  = 3887404748
	EXO    = 898834093
)

// Hash values for Gender 'genderHash' JSON key
const (
	MALE   = 3111576190
	FEMALE = 2204441813
)

// Gender Enum values used in some of the Bungie API responses
const (
	MaleEnum          = 0
	FemaleEnum        = 1
	UnknownGenderEnum = 2
)

// BungieMembershipType constant values
const (
	XBOX     = uint(1)
	PSN      = uint(2)
	BLIZZARD = uint(4)
	DEMON    = uint(10)
)

// DamageType will represent the specific damage type tied to a weapon or NoDamage if an item does
// not deal any damage.
type DamageType uint

const (
	// NoDamage is the default and indicates an item is not a weapon
	NoDamage DamageType = 0
	// KineticDamage is tied to kinetic weapons
	KineticDamage DamageType = 3373582085
	// ArcDamage indicates a weapon does Arc damage
	ArcDamage DamageType = 2303181850
	// VoidDamage indicates a weapon deals Void damage
	VoidDamage DamageType = 3454344768
	// SolarDamage indicates a weapon does Solar Damage
	SolarDamage DamageType = 1847026933
)

// Progression type hashes
const (
	valorHash  = "3882308435"
	gloryHash  = "2679551909"
	infamyHash = "2772425241"
)

var (
	pvpRankSteps = [6]string{
		"Guardian",
		"Brave",
		"Heroic",
		"Fabled",
		"Mythic",
		"Legend",
	}
	gambitRankSteps = [16]string{
		"Guardian 1",
		"Guardian 2",
		"Guardian 3",
		"Brave 1",
		"Brave 2",
		"Brave 3",
		"Heroic 1",
		"Heroic 2",
		"Heroic 3",
		"Fabled 1",
		"Fabled 2",
		"Fabled 3",
		"Mythic 1",
		"Mythic 2",
		"Mythic 3",
		"Legend",
	}
)

// Alexa doesn't understand some of the dsetiny items or splits them into separate words
// This will allow us to translate to the correct name before doing the lookup.
var commonAlexaItemTranslations = map[string]string{
	"spin metal":        "spinmetal",
	"spin mental":       "spinmetal",
	"passage coins":     "passage coin",
	"strange coins":     "strange coin",
	"exotic shards":     "exotic shard",
	"worm spore":        "wormspore",
	"3 of coins":        "three of coins",
	"worms for":         "wormspore",
	"worm for":          "wormspore",
	"motes":             "mote of light",
	"motes of light":    "mote of light",
	"spin middle":       "spinmetal",
	"e d.s that tokens": "edz token",
	"e.d.z token":       "edz token",
	"e.d.z tokens":      "edz token",
	"edz tokens":        "edz token",
	"edc token":         "edz token",
	"edc tokens":        "edz token",
}

var commonAlexaClassNameTrnaslations = map[string]string{
	"fault": "vault",
	"tatum": "titan",
}
