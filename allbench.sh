NUM=$1
PLL=$2
if [ -z "$1" ]; then
    NUM=200
fi
if [ -z "$2" ]; then
    PLL=$((NUM/20))
fi
printf "\n** JSON plain **\n" > allbench.txt
weighttp -n $NUM http://gospel99.herokuapp.com/msg/ -H "Accept: application/json" -c $PLL -t 2 >> allbench.txt
printf "\n**JSON gzip **\n" >> allbench.txt
weighttp -n $NUM http://gospel99.herokuapp.com/msg/ -H "Accept: application/json" -H "Accept-encoding: gzip" -c $PLL -t 2 >> allbench.txt
printf "\n**msgpack plain **\n" >> allbench.txt
weighttp -n $NUM http://gospel99.herokuapp.com/msg/ -H "Accept: application/x-msgpack" -c $PLL -t 2 >> allbench.txt
printf "\n**msgpack gzip **\n" >> allbench.txt
weighttp -n $NUM http://gospel99.herokuapp.com/msg/ -H "Accept: application/x-msgpack" -H "Accept-encoding: gzip" -c $PLL -t 2 >> allbench.txt
printf "\n**home plain **\n" >> allbench.txt
weighttp -n $NUM http://gospel99.herokuapp.com/ -c $PLL -t 2 >> allbench.txt
printf "\n**home gzip **\n" >> allbench.txt
weighttp -n $NUM http://gospel99.herokuapp.com/ -H "Accept-encoding: gzip" -c $PLL -t 2 >> allbench.txt
printf "\n\n"
cat allbench.txt | egrep "\*\*|finished"