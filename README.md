# edf-importer
Tool to import .edf files to influxdb or victoriametrics.
I created this tool to import data from my ResMed AirSense 11 AutoSet CPAP machine into VictoriaMetrics.
It may not necessarily correctly import data from other models or brands yet. 
It does support both timeseries metrics and annotations (events).

## Caveats
- The path is expected to be a directory filled with datestamped dirs in `YYYYmmdd` format.
- The filenames to read are currently requiring to end in BRP.edf, EVE.edf, or PLD.edf.
- EDF / CPAP data is usually sampled at a very high rate. Make sure your database is configured
to allow this. For Victoriametrics, for example, make sure the flag `-dedup.minScrapeInterval` is
set to a low value, in my case less than 40ms.
- The data is read with your computer's local timezone, as EDF files do not necessarily have a timezone
associated with their data.
- Zero-duration events from EDF+ annotations are automatically given a 10 second duration.
This was done to allow easier graphing of ResMed Hypopnea events.

## Running
TBD

## Details
TBD
- metric names and labels
- annotation points

## Dashboard
TBD
