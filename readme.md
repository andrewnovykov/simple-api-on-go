```
APP IN CLOUD
https://simple-api-on-go-redstudio-da809bd3.koyeb.app/

BUILD APP
go build -o myapp

CREATE USER
curl -X POST -H "Content-Type: application/json" -d '{"email":"your-email@example.com", "password":"your-password"}' http://localhost:8080/register

LOGIN
curl -X POST -H "Content-Type: application/json" -d '{"email":"your-email@example.com", "password":"your-password"}' http://localhost:8080/login

CREATE ITEM
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer your-token" -d '{"name":"item-name", "price":10.5}' http://localhost:8080/items

UPDATE ITEM
curl -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer your-token" -d '{"ids":[2,1],"item":{"name":"New Item"}}' http://localhost:8080/updateitems

GET ITEMS
curl http://localhost:8080/items

GET ITEM
curl http://localhost:8080/items/1

```
