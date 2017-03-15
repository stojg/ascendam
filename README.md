# Ascendam

My own personal web site up / down checker.
 
I'm using it for testing and developing deployment tools and to measure eventual down time during 
server / application switch over.

> From latin: ad- (“[up] to”) +‎ scandō (“climb”).

## Usage

	ascendam -url=http://twitter.com -timeout=1

Will constantly check if http://twitter.com responds with a 200 status within 1 second. HTTP redirection is 
regarded as down.

Example output:

	$ $ ascendam -url=https://twitter.com -timeout=1
	Running uptime check on 'https://twitter.com'
	Timeout is set to 1s and pause 1s between checks
	Stop ascendam with ctrl+c
      
	2017/03/15 22:24:31 Up	200	629.559281ms
	2017/03/15 22:24:39 Down	n/a	975.512435ms	request timed out
	2017/03/15 22:24:45 Up	200	502.761937ms
	^C
	6 outages of 16 checks 
	Average loadtime: 490.592418ms 
	Downtime: 6.003754117s 
	Uptime: 10.491571835s

## Installation via go get

To install ascendam, please use go get. I tag versions so feel free to checkout that tag and recompile.

	$ go get github.com/stojg/ascendam
	...
	$ ascendam -h

## Installation via binaries

Check [https://github.com/stojg/ascendam/releases](https://github.com/stojg/ascendam/releases) if there is a pre-built 
binary for your platform. If there is download it, unzip and put somewhere in your path.




