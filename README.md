# ITPE-report
ITPE report tool - Used to fetch power info from existing exporter and generate report for perf &amp; energy

## Usage

### Host
1. Compile with `make build`
2. Check help message with `./bin/itpe-report --help`
3. Run the application with `./bin/itpe-report`

### Docker
1. `docker build -t itpe-report .`
2. Check help message with `docker run --rm itpe-report --help`
3. Run the application with `docker run --rm itpe-report <args>`
