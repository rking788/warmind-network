package bungie

const (
	solarSingeHash uint = 2558957669
	arcSingeHash   uint = 3215384520
	voidSingeHash  uint = 3362074814
)

// Destiny.TierType
const (
	unknownTier  = 0
	currencyTier = 1
	basicTier    = 2
	commonTier   = 3
	rareTier     = 4
	superiorTier = 5
	exoticTier   = 6
)

// Hash values for different class types 'classHash' JSON key
const (
	warlock = 2271682572
	titan   = 3655393761
	hunter  = 671679327
)

// SubclassHash holds the hash representing the exact subclass "item". These are represented with
// items and a subclass bucket definition.
type SubclassHash uint

// Subclass hashes (from the subclass inventory bucket)
const (
	striker    SubclassHash = 2958378809
	sunbreaker SubclassHash = 3105935002
	sentinel   SubclassHash = 3382391785

	srcstrider   SubclassHash = 1334959255
	gunslinger   SubclassHash = 3635991036
	nightstalker SubclassHash = 3225959819

	stormcaller SubclassHash = 1751782730
	dawnblade   SubclassHash = 3481861797
	voidwalker  SubclassHash = 3887892656
)

// Hash values for Race types 'raceHash' JSON key
const (
	awoken = 2803282938
	human  = 3887404748
	exo    = 898834093
)

// Hash values for Gender 'genderHash' JSON key
const (
	male   = 3111576190
	female = 2204441813
)

// DamageType will represent the specific damage type tied to a weapon or NoDamage if an item does
// not deal any damage.
type DamageType uint

const (
	// NoDamage is the default and indicates an item is not a weapon
	noDamage DamageType = 0
	// KineticDamage is tied to kinetic weapons
	kineticDamage DamageType = 3373582085
	// ArcDamage indicates a weapon does Arc damage
	arcDamage DamageType = 2303181850
	// VoidDamage indicates a weapon deals Void damage
	voidDamage DamageType = 3454344768
	// SolarDamage indicates a weapon does Solar Damage
	solarDamage DamageType = 1847026933
)

// Progression type hashes
const (
	valorHash  = "3882308435"
	gloryHash  = "2679551909"
	infamyHash = "2772425241"
)

const (
	kineticBucket = 1498876634
	energyBucket  = 2465295065
	powerBucket   = 953998645
	ghostBucket   = 4023194814

	helmetBucket     = 3448274439
	gauntletsBucket  = 3551918588
	chestBucket      = 14239492
	legsBucket       = 20886954
	artifactBucket   = 434908299
	classArmorBucket = 1585787867
	subclassBucket   = 3284755031
)
