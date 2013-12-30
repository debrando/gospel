for enc in json x-bson x-msgpack; do
	for c in "" "--compress"; do
		NAME=/tmp/test$enc$c.out
		TRASF=`curl -X GET http://gospel99.herokuapp.com/msg/ -o $NAME -H "Accept: application/$enc" $c -v 2>&1 | grep '\-\-' | tail -n 1 | cut -d ' ' -f 2`
		BYTES=`ls -lart $NAME | cut -d ' ' -f 8`
		echo " $TRASF -> $BYTES | $enc $c"
	done
done
