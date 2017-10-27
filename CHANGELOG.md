Version 1.3.0
===============
- Now correctly checking all last played dates to find the most recent Destiny membership
  instead of just using the first one returned in the response
- Added a new filter to check item required level values to avoid trying to equip weapons
  that require a higher level
- Refactored the Bungie client to use a simplified request flow and consolidate retry logic
- Renamed everything from guardian helper to warmind network
- Removed the unused push script

Version 1.2.0
===============
- Named loadout support
- Auto deploying from Travis CI for staging branches

Version 1.0.0
===============
- Updated to support the Destiny 2 API.
- Added the DestinyJoke intent handler to tell random Destiny related jokes.
- Refactored the way intent handlers are defined.
- Added a AuthHandler middleware to wrap calls that require access_tokens from Bungie.

Version 0.3.0
===============
- Added support for unloading engrams to the vault
- Added support for equipping max light loadouts to the current character

Version 0.2.0
===============
- Added support for querying data from [Trials Report](https://trials.report) ([Github](https://github.com/DestinyTrialsReport/DestinyTrialsReport))
  - Top used weapons and categories
  - Personal performance stats for current week
  - Personal top used weapons
  - Current map

Version 0.1.0
===============
- Initial "minimum viable product" release.
