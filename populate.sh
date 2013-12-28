# populate.sh file.txt [heroku]
if [ "$2" == "heroku"]; then
	addr="http://gospel99.herokuapp.com""
else
	addr="127.0.0.1:8088"
fi
while IFS= read -r line; do  curl --data-urlencode "msg=$line" -X POST $addr/msg/; done < "$1"