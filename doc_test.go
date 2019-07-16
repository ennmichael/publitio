// Package publitio provides Go language API for the https://publit.io site.
package publitio

import (
	"net/url"
	"os"
)

func Example() {
	api := API{Key: "xxx", Secret: "yyy"}
	api.Get("files/list", url.Values{"limit": {"12"}}) // List at most 12 files
	api.Get("files/list", url.Values{})                // List all files, no query parameters
	api.Get("/files/list", nil)                        // You can always pass nil if you have no query parameters

	reader, _ := os.Open("path/to/file")

	// Upload a file from memory and give it a title
	api.UploadFile(reader, url.Values{"title": {"My file"}})

	// Upload a file from a remote URL and give it a custom ID
	api.UploadFile(nil, url.Values{"file_url": {"https://example.org"}, "public_id": {"xxGh332"}})

	api.Delete("files/delete/fileId", url.Values{}) // Delete a file with ID fileID
	api.Delete("/files/delete/fileId", nil)         // Same

	// For more complete documentation, see https://publit.io/docs/
}

func ExampleAPI_UploadFile() {
	api := API{Key: "xxx", Secret: "yyy"}
	reader, _ := os.Open("path/to/file")

	// Upload a file from memory and give it a title
	api.UploadFile(reader, url.Values{"title": {"My file"}})

	// Upload a file from a remote URL and give it a custom ID
	api.UploadFile(nil, url.Values{"file_url": {"https://example.org"}, "public_id": {"xxGh332"}})
}

func ExampleAPI_Delete() {
	api := API{Key: "xxx", Secret: "yyy"}
	api.Delete("files/delete/fileId", url.Values{}) // Delete a file with ID fileID
	api.Delete("/files/delete/fileId", nil)         // Same
}

func ExampleAPI_Put() {
	api := API{Key: "xxx", Secret: "yyy"}
	api.Put("files/update/fileId", url.Values{"title": {"New title"}}) // Update the title for a file with ID fileID
}

func ExampleAPI_Get() {
	api := API{Key: "xxx", Secret: "yyy"}
	api.Get("files/list", url.Values{"limit": {"12"}}) // List at most 12 files
	api.Get("files/list", url.Values{})                // List all files, no query parameters
	api.Get("/files/list", nil)                        // You can always pass nil if you have no query parameters
}
