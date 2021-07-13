db.images.createIndex( { "album": 1, "id": 1 }, { unique: true } )
db.images.createIndex( { "album": 1, "compressed": 1 } )
db.images.createIndex( { "album": 1, "rating": -1 } )
db.edges.createIndex( { "album": 1, "from": 1, "to": 1 }, { unique: true } )
