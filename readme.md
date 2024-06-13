```


curl -X POST -H "Content-Type: application/json" -d '{"name":"item1", "price":10.5}' http://localhost:8080/items

curl http://localhost:8080/items

curl http://localhost:8080/items/1

curl -X PUT -H "Content-Type: application/json" -d '{"ids":[2,1],"item":{"name":"New Item"}}' http://localhost:8080/updateitems


```
