Version 1.8.1
===============
- Added intents for requesting top activities and meta stats from Charlemagne.
- Performance improvements around filtering items from a list.
- Added the Sentry.io library for reporting unexpected errors so alerts can be provided.
- Improved documentation to silence Go linter warnings
- Fixed a bug causing users to not be able to equip a named loadout.

Version 1.7.0
===============
- Added an intent for randomizing loadouts.
- Requiring confirmation for loadout names now to avoid accidental loadout names.
- Adding support for Canada and Australia (English only still).

Version 1.6.0
===============
- No longer supporting Guardian Helper, asking users to enable Warmind Network.

Version 1.5.0
===============
- Added an intent for listing a user's loadouts that are currently persisted.
- Fixing an issue with the popular Trials weapons returning an empty string if
    the user hasn't played any matches.

Version 1.4.2
===============
- Fixed an issue reading top weapons from Trials Report.
- Now returning the game mode for Trials in the Alexa response,
  not just the map name.
- Updated the trials package to use the Trials of the Nine data from
  Trials Report instead of the legacy Trails of Osiris stuff.
- Now reading the port variable from the config or environment to allow
    Nginx to listen on port 443 and redirect to a different port for the Go apps.

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
