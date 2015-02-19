# Ascendam

My own personal website up / down checker.
 
It's good for testing deployment tools and that there is no downtime during deployment.

## Usage

	./ascendam -url=http://twitter.com -max-ms=1000

Will constantly check if http://twitter.com responds with a 200 status within 1000ms.

Example output:

	$ ./ascendam -url=http://twitter.com -max-ms=1000
	Running uptime check on 'http://twitter.com'
	Timeout is set to 1s
	12:07:04 Down 200 1.554766774s
	12:07:04 Up   200 778.494798ms


## Downloads

Check [https://github.com/stojg/ascendam/releases](https://github.com/stojg/ascendam/releases) if there is a pre-built binary for your platform.

## Build

	$ go build ascendam.go

