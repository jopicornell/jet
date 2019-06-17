package tests

import (
	"encoding/json"
	"fmt"
	. "github.com/go-jet/jet/sqlbuilder"
	"github.com/go-jet/jet/tests/.test_files/dvd_rental/chinook/model"
	. "github.com/go-jet/jet/tests/.test_files/dvd_rental/chinook/table"
	"gotest.tools/assert"
	"io/ioutil"
	"testing"
)

func TestSelect(t *testing.T) {
	stmt := Album.
		SELECT(Album.AllColumns).
		ORDER_BY(Album.AlbumId.ASC())

	assertStatementSql(t, stmt, `
SELECT "Album"."AlbumId" AS "Album.AlbumId",
     "Album"."Title" AS "Album.Title",
     "Album"."ArtistId" AS "Album.ArtistId"
FROM chinook."Album"
ORDER BY "Album"."AlbumId" ASC;
`)
	dest := []model.Album{}

	err := stmt.Query(db, &dest)

	assert.NilError(t, err)
	assert.Equal(t, len(dest), 347)
	assert.DeepEqual(t, dest[0], album1)
	assert.DeepEqual(t, dest[1], album2)
	assert.DeepEqual(t, dest[len(dest)-1], album347)
}

func TestJoinEverything(t *testing.T) {

	manager := Employee.AS("Manager")

	stmt := Artist.
		LEFT_JOIN(Album, Artist.ArtistId.EQ(Album.ArtistId)).
		LEFT_JOIN(Track, Track.AlbumId.EQ(Album.AlbumId)).
		LEFT_JOIN(Genre, Genre.GenreId.EQ(Track.GenreId)).
		LEFT_JOIN(MediaType, MediaType.MediaTypeId.EQ(Track.MediaTypeId)).
		LEFT_JOIN(PlaylistTrack, PlaylistTrack.TrackId.EQ(Track.TrackId)).
		LEFT_JOIN(Playlist, Playlist.PlaylistId.EQ(PlaylistTrack.PlaylistId)).
		LEFT_JOIN(InvoiceLine, InvoiceLine.TrackId.EQ(Track.TrackId)).
		LEFT_JOIN(Invoice, Invoice.InvoiceId.EQ(InvoiceLine.InvoiceId)).
		LEFT_JOIN(Customer, Customer.CustomerId.EQ(Invoice.CustomerId)).
		LEFT_JOIN(Employee, Employee.EmployeeId.EQ(Customer.SupportRepId)).
		LEFT_JOIN(manager, manager.EmployeeId.EQ(Employee.ReportsTo)).
		SELECT(
			Artist.AllColumns,
			Album.AllColumns,
			Track.AllColumns,
			Genre.AllColumns,
			MediaType.AllColumns,
			PlaylistTrack.AllColumns,
			Playlist.AllColumns,
			Invoice.AllColumns,
			Customer.AllColumns,
			Employee.AllColumns,
			manager.AllColumns,
		).
		ORDER_BY(Artist.ArtistId, Album.AlbumId, Track.TrackId,
			Genre.GenreId, MediaType.MediaTypeId, Playlist.PlaylistId,
			Invoice.InvoiceId, Customer.CustomerId).
		WHERE(Artist.ArtistId.LT_EQ(Int(100000)))

	var dest []struct { //list of all artist
		model.Artist

		Albums []struct { // list of albums per artist
			model.Album

			Tracks []struct { // list of tracks per album
				model.Track

				Genre     model.Genre     // track genre
				MediaType model.MediaType // track media type

				Playlists []model.Playlist `sql:"table:Playlist"` // list of playlist where track is used

				Invoices []struct { // list of invoices where track occurs
					model.Invoice

					Customer struct { // customer data for invoice
						model.Customer

						Employee *struct {
							model.Employee

							Manager *model.Employee
						}
					}
				}
			}
		}
	}

	err := stmt.Query(db, &dest)

	assert.NilError(t, err)
	//jsonSave(dest)

	fmt.Println("Artist count :", len(dest))
	assert.Equal(t, len(dest), 275)

	assertJson(t, "./data/joined_everything.json", dest)
}

func TestUnionForQuotedNames(t *testing.T) {

	stmt := UNION_ALL(
		Album.SELECT(Album.AllColumns).WHERE(Album.AlbumId.EQ(Int(1))),
		Album.SELECT(Album.AllColumns).WHERE(Album.AlbumId.EQ(Int(2))),
	).
		ORDER_BY(Album.AlbumId)

	fmt.Println(stmt.DebugSql())
	assertStatementSql(t, stmt, `
(
     (
          SELECT "Album"."AlbumId" AS "Album.AlbumId",
               "Album"."Title" AS "Album.Title",
               "Album"."ArtistId" AS "Album.ArtistId"
          FROM chinook."Album"
          WHERE "Album"."AlbumId" = 1
     )
     UNION ALL
     (
          SELECT "Album"."AlbumId" AS "Album.AlbumId",
               "Album"."Title" AS "Album.Title",
               "Album"."ArtistId" AS "Album.ArtistId"
          FROM chinook."Album"
          WHERE "Album"."AlbumId" = 2
     )
)
ORDER BY "Album.AlbumId";
`, int64(1), int64(2))

	dest := []model.Album{}

	err := stmt.Query(db, &dest)

	assert.NilError(t, err)

	assert.Equal(t, len(dest), 2)
	assert.DeepEqual(t, dest[0], album1)
	assert.DeepEqual(t, dest[1], album2)
}

func assertJson(t *testing.T, jsonFilePath string, data interface{}) {
	fileJsonData, err := ioutil.ReadFile(jsonFilePath)
	assert.NilError(t, err)

	jsonData, err := json.MarshalIndent(data, "", "\t")
	assert.NilError(t, err)

	assert.Assert(t, string(fileJsonData) == string(jsonData))
}

func jsonPrint(v interface{}) {
	json, _ := json.MarshalIndent(v, "", "\t")
	fmt.Println(string(json))
}

func jsonSave(path string, v interface{}) {
	json, _ := json.MarshalIndent(v, "", "\t")

	err := ioutil.WriteFile(path, json, 0644)

	if err != nil {
		panic(err)
	}
}

var album1 = model.Album{
	AlbumId:  1,
	Title:    "For Those About To Rock We Salute You",
	ArtistId: 1,
}

var album2 = model.Album{
	AlbumId:  2,
	Title:    "Balls to the Wall",
	ArtistId: 2,
}

var album347 = model.Album{
	AlbumId:  347,
	Title:    "Koyaanisqatsi (Soundtrack from the Motion Picture)",
	ArtistId: 275,
}