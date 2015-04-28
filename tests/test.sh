#!/bin/sh
proxy=localhost:8080
curl -s -r 0-20000 http://s0.cyberciti.org/images/misc/static/2012/11/ifdata-welcome-0.png -o part1 -x $proxy
curl -s -r 20001-36907 http://s0.cyberciti.org/images/misc/static/2012/11/ifdata-welcome-0.png -o part2 -x $proxy
cat part1 part2 > test2.png

if ! cmp test.png test2.png >/dev/null 2>&1
then
    echo Test failed: test2.png != test.png
fi
rm part?
