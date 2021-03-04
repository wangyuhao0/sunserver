module sunserver

go 1.15

require (
	github.com/duanhf2012/origin v0.0.0-20210125063345-0cfb2163463c
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.4.3
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

//replace github.com/duanhf2012/origin => ./origin
