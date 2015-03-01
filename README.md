# Ascendam

My own personal web site up / down checker.
 
I'm personally using it for testing deployment tools and to measure eventual down time.

## Usage

	ascendam -url=http://twitter.com -max-ms=1300

Will constantly check if http://twitter.com responds with a 200 status within 1300ms. HTTP redirection is 
regarded as down.

Example output:

	$ ascendam -url=https://twitter.com -max-ms=1300
    Running uptime check on 'https://twitter.com'
    Timeout is set to 1.3s
    15:29:53 Down	n/a	1.305579983s	request timed out (cancelled)
    15:29:56 Up		200	897.269721ms
    15:29:58 Down	n/a	1.304362985s	request timed out (cancelled)

## Installation via go get

To install ascendam, please use go get. I tag versions so feel free to checkout that tag and recompile.

	$ go get github.com/stojg/ascendam
	...
	$ ascendam -h

## Installation via binaries

Check [https://github.com/stojg/ascendam/releases](https://github.com/stojg/ascendam/releases) if there is a pre-built 
binary for your platform. If there is download it, unzip and put somewhere in your path.




