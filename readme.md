```

https://simple-api-on-go-redstudio-da809bd3.koyeb.app/

curl -X POST -H "Content-Type: application/json" -d '{"email":"your-email@example.com", "password":"your-password"}' http://localhost:8080/register

curl -X POST -H "Content-Type: application/json" -d '{"email":"your-email@example.com", "password":"your-password"}' http://localhost:8080/login

curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer your-token" -d '{"name":"item-name", "price":10.5}' http://localhost:8080/items

curl http://localhost:8080/items

curl http://localhost:8080/items/1

curl -X PUT -H "Content-Type: application/json" -d '{"ids":[2,1],"item":{"name":"New Item"}}' http://localhost:8080/updateitems

curl -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer your-token" -d '{"ids":[2,1],"item":{"name":"New Item"}}' http://localhost:8080/updateitems

go build -o myapp

curl -X POST -H "Content-Type: application/json" -d '{"name":"item1", "price":10.5}' https://simple-api-on-go-redstudio-da809bd3.koyeb.app/items


```
