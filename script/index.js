var conn = new Mongo("mongodb://admin:123456@49.232.105.112:27017/admin");
var db = conn.getDB("testdb");
var checkServers=db.getCollection('userinfo').createIndex({key1:1})
