Version 1.4.2
===============
- Fixed an issue reading top weapons from Trials Report.
- Now returning the game mode for Trials in the Alexa response, 
  not just the map name.
- Updated the trials package to use the Trials of the Nine data from
  Trials Report instead of the legacy Trails of Osiris stuff.
- Now reading the port variable from the config or environment to allow
    Nginx to listen on port 443 and redirect to a different port for the Go apps.
