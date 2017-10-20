Version 1.3.0
===============
- Now correctly checking all last played dates to find the most recent Destiny membership
  instead of just using the first one returned in the response
- Added a new filter to check item required level values to avoid trying to equip weapons
  that require a higher level
- Refactored the Bungie client to use a simplified request flow and consolidate retry logic
- Renamed everything from guardian helper to warmind network
- Removed the unused push script

