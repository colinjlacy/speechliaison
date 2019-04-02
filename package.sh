#!

rm -f ./deploy/speechLiason.zip
GOOS=linux go build -o ./deploy/speechLiason
cd ./deploy/
zip speechLiason.zip speechLiason
cd ../
aws lambda update-function-code --function-name reborneScanHandler \
--zip-file fileb://deploy/speechLiason.zip