# Purpose
This tool will create tags against PANOS (tested on PANOS v8.1).

## Features
- Creates tags loaded from file
- Built-in concurrency (levarages go-routines to process X # of PAN devices)
- Attempts to create tag 3x before confirming device unreachable

## Usage
1. Store API Key in local .env file - example provided in repo
   ```
   API_KEY=LUFRPT14MW5xOEo1R09KVlBZNnpnemh0VHRBOWl6TGM9bXcwM3JHUGVhRlNiY0dCR0srNERUQT09
   PAN=192.168.1.35
   ```
2. Pass in filename that has desired tags to be created (one per line)
   
   `./prog <filename>`

## Authors 
- Bobby Williams <https://www.linkedin.com/in/bobby-williams-48222450/>