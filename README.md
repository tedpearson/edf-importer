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

Rename `sdf-importer.example.ymal` to `sdf-importer.yaml` and fill in your database config.

Sdf-importer will continue to run and watch the specified path, importing new data the next time
you insert your SD card.

    ./edf-importer --config sdf-importer.yaml --path /Volumes/<drivename>/folder    

    Usage of ./edf-importer:
      -config string
        	Config file (default "sdf-importer.yaml")
      -dry-run
        	Don't insert into the database
      -path string
        	Path to data directory (default "/Volumes/NO NAME/DATALOG")
      -state-file string
        	State file (default "edf-importer.state.yaml")
      -v	Show version and exit

## Details

Info about exported data:

Metric name: **cpap**
- fields:
  - "edf supplied name"
    - name will be whatever the EDF label name is from the EDF file.
    - contains the raw data from the EDF file
  - annotation
    - value of 1 during the event
    - value of 0 before and after the event
- tags
  - event
    - contains the EDF+ annotation name
    - only on annotation events, not raw data

## Dashboard
TBD
