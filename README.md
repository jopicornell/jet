# Jet

[![CircleCI](https://circleci.com/gh/go-jet/jet/tree/develop.svg?style=svg&circle-token=97f255c6a4a3ab6590ea2e9195eb3ebf9f97b4a7)](https://circleci.com/gh/go-jet/jet/tree/develop)

Jet is a framework for writing type-safe SQL queries for PostgreSQL in Go, with ability to easily 
convert database query result to desired arbitrary structure.  
_Support for additional databases will be added in future jet releases._


## Contents
   - [Features](#features)
   - [Getting Started](#getting-started)
      - [Prerequisites](#prerequisites)
      - [Installation](#installation)
      - [Quick Start](#quick-start)
         - [Generate sql builder and model files](#generate-sql-builder-and-model-files)
         - [Lets write some SQL queries in Go](#lets-write-some-sql-queries-in-go)
         - [Execute query and store result](#execute-query-and-store-result)
   - [Benefits](#benefits)
   - [Dependencies](#dependencies)
   - [Versioning](#versioning)

## Features
 1) Type-safe SQL Builder
     - Types - boolean, integers(smallint, integer, bigint), floats(real, numeric, decimal, double precision), 
     strings(text, character, character varying), date, time(z), timestamp(z) and enums.
     - Statements:
        * SELECT (DISTINCT, FROM, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT, OFFSET, FOR, UNION, INTERSECT, EXCEPT, sub-queries) 
        * INSERT (VALUES, query, RETURNING), 
        * UPDATE (SET, WHERE, RETURNING), 
        * DELETE (WHERE, RETURNING),
        * LOCK (IN, NOWAIT)
 2) Auto-generated Data Model types - Go struct mapped to database type (table or enum)
 3) Query execution with mapping to arbitrary destination structure - destination structure can be 
 created by combining auto-generated data model types. 

## Getting Started

### Prerequisites

To install Jet package, you need to install Go and set your Go workspace first.

[Go](https://golang.org/) **version 1.8+ is required**

### Installation

Use the bellow command to install jet
```sh
$ go get -u github.com/go-jet/jet
```

Install jet generator to GOPATH bin folder. This will allow generating jet files from the command line.

```sh
go install github.com/go-jet/jet/cmd/jet
```

Make sure GOPATH bin folder is added to the PATH environment variable.

### Quick Start
For this quick start example we will use sample _dvd rental_ database. Full database dump can be found in [./tests/init/data/dvds.sql](./tests/init/data/dvds.sql).
Schema diagram of interest for example can be found [here](./examples/quick-start/diagram.png).

#### Generate sql builder and model files
To generate jet SQL Builder and Data Model files from postgres database we need to call `jet` generator with postgres 
connection parameters and root destination folder path for generated files.\
Assuming we are running local postgres database, with user `jet`, database `jetdb` and schema `dvds` we will use this command:
```sh
jet -host=localhost -port=5432 -user=jet -password=jet -dbname=jetdb -schema=dvds -path=./gen
```
```sh
Connecting to postgres database: host=localhost port=5432 user=jet password=jet dbname=jetdb sslmode=disable 
Retrieving schema information...
    FOUND 15  table(s),  1  enum(s)
Destination directory: ./gen/jetdb/dvds
Cleaning up schema destination directory...
Generating table sql builder files...
Generating table model files...
Generating enum sql builder files...
Generating enum model files...
Done
```

As command output suggest, Jet will:
- connect to postgres database and retrieve information about the _tables_ and _enums_ of `dvds` schema
- delete everything in schema destination folder -  `./gen/jetdb/dvds`,   
- and finally generate sql builder and model files for each schema tables and enums.  


Generated files folder structure will look like this:
```sh 
|-- gen                               # -path
|   `-- jetdb                         # database name
|       `-- dvds                      # schema name
|           |-- enum                  # sql builder folder for enums
|           |   |-- mpaa_rating.go
|           |-- table                 # sql builder folder for tables
|               |-- actor.go
|               |-- address.go
|               |-- category.go
|               ...
|           |-- model                 # Plain Old Data for every table and enum
|           |   |-- actor.go
|           |   |-- address.go
|           |   |-- mpaa_rating.go
|           |   ...
```
Types from `table` and `enum` are used to write type safe SQL in Go, and `model` types are combined to store 
results of the SQL queries.

#### Lets write some SQL queries in Go

First we need to import jet and generated files from previous step:
```go
import (
	// dot import so that Go code would resemble as much as native SQL
	// dot import is not mandatory
	. "github.com/go-jet/jet"                                           
	. "github.com/go-jet/jet/examples/quick-start/gen/jetdb/dvds/table" 

	"github.com/go-jet/jet/examples/quick-start/gen/jetdb/dvds/model"
)
```
Lets say we want to retrieve the list of all _actors_ that acted in _films_ longer than 180 minutes, _film language_ is 'English' 
and _film category_ is not 'Action'.  
```go
stmt := SELECT(
    Actor.ActorID, Actor.FirstName, Actor.LastName, Actor.LastUpdate,  // or just Actor.AllColumns
    Film.AllColumns,                                                  
    Language.AllColumns,
    Category.AllColumns,
).FROM(
    Actor.
        INNER_JOIN(FilmActor, Actor.ActorID.EQ(FilmActor.ActorID)).  
        INNER_JOIN(Film, Film.FilmID.EQ(FilmActor.FilmID)).          
        INNER_JOIN(Language, Language.LanguageID.EQ(Film.LanguageID)).
        INNER_JOIN(FilmCategory, FilmCategory.FilmID.EQ(Film.FilmID)).
        INNER_JOIN(Category, Category.CategoryID.EQ(FilmCategory.CategoryID)),
).WHERE(
    Language.Name.EQ(String("English")).             
        AND(Category.Name.NOT_EQ(String("Action"))).  
        AND(Film.Length.GT(Int(180))),               
).ORDER_BY(
    Actor.ActorID.ASC(),
    Film.FilmID.ASC(),
)
```
With package(dot) import above statements looks almost the same as native SQL.  
Note that every column has a type. String column `Language.Name` and `Category.Name` can be compared only with 
string columns and expressions. `Actor.ActorID`, `FilmActor.ActorID`, `Film.Length` are integer columns 
and can be compared only with integer columns and expressions.

__How get parametrized SQL query?__
```go
query, args, err := stmt.Sql()
```
query - parametrized query\
args - parameters for the query

<details>
  <summary>Click to see `query` and `arg`</summary>
  
```sql
SELECT actor.actor_id AS "actor.actor_id",
     actor.first_name AS "actor.first_name",
     actor.last_name AS "actor.last_name",
     actor.last_update AS "actor.last_update",
     film.film_id AS "film.film_id",
     film.title AS "film.title",
     film.description AS "film.description",
     film.release_year AS "film.release_year",
     film.language_id AS "film.language_id",
     film.rental_duration AS "film.rental_duration",
     film.rental_rate AS "film.rental_rate",
     film.length AS "film.length",
     film.replacement_cost AS "film.replacement_cost",
     film.rating AS "film.rating",
     film.last_update AS "film.last_update",
     film.special_features AS "film.special_features",
     film.fulltext AS "film.fulltext",
     language.language_id AS "language.language_id",
     language.name AS "language.name",
     language.last_update AS "language.last_update",
     category.category_id AS "category.category_id",
     category.name AS "category.name",
     category.last_update AS "category.last_update"
FROM dvds.actor
     INNER JOIN dvds.film_actor ON (actor.actor_id = film_actor.actor_id)
     INNER JOIN dvds.film ON (film.film_id = film_actor.film_id)
     INNER JOIN dvds.language ON (language.language_id = film.language_id)
     INNER JOIN dvds.film_category ON (film_category.film_id = film.film_id)
     INNER JOIN dvds.category ON (category.category_id = film_category.category_id)
WHERE ((language.name = $1) AND (category.name != $2)) AND (film.length > $3)
ORDER BY actor.actor_id ASC, film.film_id ASC;
```
```sh 
[English Action 180]
```


</details>
    
__How to get debug SQL that can be copy pasted to sql editor and executed?__
```go
debugSql, err := stmt.DebugSql()
```
debugSql - parametrized query where every parameter is replaced with appropriate string argument representation
<details>
  <summary>Click to see debug sql</summary>
  
```sql
SELECT actor.actor_id AS "actor.actor_id",
     actor.first_name AS "actor.first_name",
     actor.last_name AS "actor.last_name",
     actor.last_update AS "actor.last_update",
     film.film_id AS "film.film_id",
     film.title AS "film.title",
     film.description AS "film.description",
     film.release_year AS "film.release_year",
     film.language_id AS "film.language_id",
     film.rental_duration AS "film.rental_duration",
     film.rental_rate AS "film.rental_rate",
     film.length AS "film.length",
     film.replacement_cost AS "film.replacement_cost",
     film.rating AS "film.rating",
     film.last_update AS "film.last_update",
     film.special_features AS "film.special_features",
     film.fulltext AS "film.fulltext",
     language.language_id AS "language.language_id",
     language.name AS "language.name",
     language.last_update AS "language.last_update",
     category.category_id AS "category.category_id",
     category.name AS "category.name",
     category.last_update AS "category.last_update"
FROM dvds.actor
     INNER JOIN dvds.film_actor ON (actor.actor_id = film_actor.actor_id)
     INNER JOIN dvds.film ON (film.film_id = film_actor.film_id)
     INNER JOIN dvds.language ON (language.language_id = film.language_id)
     INNER JOIN dvds.film_category ON (film_category.film_id = film.film_id)
     INNER JOIN dvds.category ON (category.category_id = film_category.category_id)
WHERE ((language.name = 'English') AND (category.name != 'Action')) AND (film.length > 180)
ORDER BY actor.actor_id ASC, film.film_id ASC;
```
</details>


#### Execute query and store result

Well formed SQL is just a first half the job. Lets see how can we make some sense of result set returned executing 
above statement. Usually this is the most complex and tedious work, but with Jet it is the easiest.

First we have to create desired structure to store query result set. 
This is done be combining autogenerated model types or it can be done manually(see wiki for more information). 

Let's say this is our desired structure:  
```go
var dest []struct {
    model.Actor
    Films []struct {
        model.Film
        Language model.Language
        Categories []model.Category
    }
}
```
_There is no limitation for how big or nested destination structure can be._

Now lets execute a above statement on open database connection db and store result into `dest`.

```go
err := stmt.Query(db, &dest)
handleError(err)
```

__And thats it.__
  
`dest` now contains the list of all actors(with list of films acted, where each film has information about language and list of belonging categories) that acted in films longer than 180 minutes, film language is 'English' 
and film category is not 'Action'.

Lets print `dest` as a json to see:
```go
jsonText, _ := json.MarshalIndent(dest, "", "\t")
fmt.Println(string(jsonText))
```

```js
[
	{
		"ActorID": 1,
		"FirstName": "Penelope",
		"LastName": "Guiness",
		"LastUpdate": "2013-05-26T14:47:57.62Z",
		"Films": [
			{
				"FilmID": 499,
				"Title": "King Evolution",
				"Description": "A Action-Packed Tale of a Boy And a Lumberjack who must Chase a Madman in A Baloon",
				"ReleaseYear": 2006,
				"LanguageID": 1,
				"RentalDuration": 3,
				"RentalRate": 4.99,
				"Length": 184,
				"ReplacementCost": 24.99,
				"Rating": "NC-17",
				"LastUpdate": "2013-05-26T14:50:58.951Z",
				"SpecialFeatures": "{Trailers,\"Deleted Scenes\",\"Behind the Scenes\"}",
				"Fulltext": "'action':5 'action-pack':4 'baloon':21 'boy':10 'chase':16 'evolut':2 'king':1 'lumberjack':13 'madman':18 'must':15 'pack':6 'tale':7",
				"Language": {
					"LanguageID": 1,
					"Name": "English             ",
					"LastUpdate": "2006-02-15T10:02:19Z"
				},
				"Categories": [
					{
						"CategoryID": 8,
						"Name": "Family",
						"LastUpdate": "2006-02-15T09:46:27Z"
					}
				]
			}
		]
	},
	{
		"ActorID": 3,
		"FirstName": "Ed",
		"LastName": "Chase",
		"LastUpdate": "2013-05-26T14:47:57.62Z",
		"Films": [
			{
				"FilmID": 996,
				"Title": "Young Language",
				"Description": "A Unbelieveable Yarn of a Boat And a Database Administrator who must Meet a Boy in The First Manned Space Station",
				"ReleaseYear": 2006,
				"LanguageID": 1,
				"RentalDuration": 6,
				"RentalRate": 0.99,
				"Length": 183,
				"ReplacementCost": 9.99,
				"Rating": "G",
				"LastUpdate": "2013-05-26T14:50:58.951Z",
				"SpecialFeatures": "{Trailers,\"Behind the Scenes\"}",
				"Fulltext": "'administr':12 'boat':8 'boy':17 'databas':11 'first':20 'languag':2 'man':21 'meet':15 'must':14 'space':22 'station':23 'unbeliev':4 'yarn':5 'young':1",
				"Language": {
					"LanguageID": 1,
					"Name": "English             ",
					"LastUpdate": "2006-02-15T10:02:19Z"
				},
				"Categories": [
					{
						"CategoryID": 6,
						"Name": "Documentary",
						"LastUpdate": "2006-02-15T09:46:27Z"
					}
				]
			}
		]
	},
	//...(125 more items)
]
```

What if, we also want to have list of films per category and actors per category, where films are longer than 180 minutes, film language is 'English' 
and film category is not 'Action'.  
In that case we can reuse above statement `stmt`, and just change our destination:

```go
var dest2 []struct {
    model.Category

    Films []model.Film
    Actors []model.Actor
}

err = stmt.Query(db, &dest2)
handleError(err)
```
<details>
  <summary>Click to see `dest2` json</summary>

```js
[
	{
		"CategoryID": 8,
		"Name": "Family",
		"LastUpdate": "2006-02-15T09:46:27Z",
		"Films": [
			{
				"FilmID": 499,
				"Title": "King Evolution",
				"Description": "A Action-Packed Tale of a Boy And a Lumberjack who must Chase a Madman in A Baloon",
				"ReleaseYear": 2006,
				"LanguageID": 1,
				"RentalDuration": 3,
				"RentalRate": 4.99,
				"Length": 184,
				"ReplacementCost": 24.99,
				"Rating": "NC-17",
				"LastUpdate": "2013-05-26T14:50:58.951Z",
				"SpecialFeatures": "{Trailers,\"Deleted Scenes\",\"Behind the Scenes\"}",
				"Fulltext": "'action':5 'action-pack':4 'baloon':21 'boy':10 'chase':16 'evolut':2 'king':1 'lumberjack':13 'madman':18 'must':15 'pack':6 'tale':7"
			},
			{
				"FilmID": 50,
				"Title": "Baked Cleopatra",
				"Description": "A Stunning Drama of a Forensic Psychologist And a Husband who must Overcome a Waitress in A Monastery",
				"ReleaseYear": 2006,
				"LanguageID": 1,
				"RentalDuration": 3,
				"RentalRate": 2.99,
				"Length": 182,
				"ReplacementCost": 20.99,
				"Rating": "G",
				"LastUpdate": "2013-05-26T14:50:58.951Z",
				"SpecialFeatures": "{Commentaries,\"Behind the Scenes\"}",
				"Fulltext": "'bake':1 'cleopatra':2 'drama':5 'forens':8 'husband':12 'monasteri':20 'must':14 'overcom':15 'psychologist':9 'stun':4 'waitress':17"
			}
		],
		"Actors": [
			{
				"ActorID": 1,
				"FirstName": "Penelope",
				"LastName": "Guiness",
				"LastUpdate": "2013-05-26T14:47:57.62Z"
			},
			{
				"ActorID": 20,
				"FirstName": "Lucille",
				"LastName": "Tracy",
				"LastUpdate": "2013-05-26T14:47:57.62Z"
			},
			{
				"ActorID": 36,
				"FirstName": "Burt",
				"LastName": "Dukakis",
				"LastUpdate": "2013-05-26T14:47:57.62Z"
			},
			{
				"ActorID": 70,
				"FirstName": "Michelle",
				"LastName": "Mcconaughey",
				"LastUpdate": "2013-05-26T14:47:57.62Z"
			},
			{
				"ActorID": 118,
				"FirstName": "Cuba",
				"LastName": "Allen",
				"LastUpdate": "2013-05-26T14:47:57.62Z"
			},
			{
				"ActorID": 187,
				"FirstName": "Renee",
				"LastName": "Ball",
				"LastUpdate": "2013-05-26T14:47:57.62Z"
			},
			{
				"ActorID": 198,
				"FirstName": "Mary",
				"LastName": "Keitel",
				"LastUpdate": "2013-05-26T14:47:57.62Z"
			}
		]
	},
    //...
]
```
</details>

Complete code example can be found at [./examples/quick-start/quick-start.go](./examples/quick-start/quick-start.go)


This example represent probably the most common use case.  Detail info about additional features and use cases can be 
found at project [wiki](https://github.com/go-jet/jet/wiki) page.

## Benefits

What are the benefits of writing SQL in Go using Jet? The biggest benefit is speed.  
Speed is improved in 3 major areas:
##### Speed of development  

Writing SQL queries is much easier directly from Go, because programmer has the help of SQL code completion and SQL type safety directly in Go.
Writing code is much faster and code is more robust. Automatic scan to arbitrary structure removes a lot of headache and 
boilerplate code needed to structure database query result.  
With Jet programmer has the power of SQL but also ease of use of NoSQL. 

##### Speed of execution

Common web and database server usually are not on the same physical machine, and there is some latency between them. 
Latency can vary from 5ms to 50+ms. In majority of cases query executed on database is simple query lasting no more than 1ms.
In those cases web server handler execution time is directly proportional to latency between server and database.
This is not such a big problem if handler calls database couple of times, but what if web server is using ORM to retrieve data from database.
ORM sometimes can access the database once for every object needed. Now lets say latency is 30ms and there are 100 
different objects required from the database. This handler will last 3 seconds !!!.  

With Jet, handler time lost on latency between server and database is constant. Because we can write complex query and 
return result in one database call. Handler execution will be proportional to the number of rows returned from database. 
ORM example replaced with jet will take just 30ms + 'result scan time' = 31ms (rough estimate).  

With Jet you can even join the whole database and store the whole structured result in  in one query call. 
This is exactly what is being done in one of the tests: [TestJoinEverything](/tests/chinook_db_test.go#L40). 
The whole test database is joined and query result is stored in a structured variable in less than 1s. 

##### How quickly bugs are found

The most expensive bugs are the one on the production and the least expensive are those found during development.
With automatically generated type safe SQL not only queries are written faster but bugs are found sooner.  
Lets return to quick start example, and take closer look at a line:
 ```go
AND(Film.Length.GT(Int(180))),
```
Lets say someone changes column `length` to `duration` from `film` table. The next go build will fail at that line and 
the bug will be caught at compile time.

Lets say someone changes the type of `length` column to some non integer type. Build will also fail at the same line
because integer columns and expressions can be only compered to other integer columns and expressions.

Without Jet these bugs will have to be either caught by some test or by manual testing. 

## Dependencies
At the moment Jet dependence only of:
- `github.com/google/uuid` _(Used for debug purposes and in data model files)_
- `github.com/lib/pq` _(Used by Jet to read information about database schema types)_

To run the tests, additional dependencies are required:
- `github.com/pkg/profile`
- `gotest.tools/assert`

## Contributing

## Versioning

[SemVer](http://semver.org/) is used for versioning. For the versions available, see the [releases](https://github.com/go-jet/jet/releases). 

## Licence