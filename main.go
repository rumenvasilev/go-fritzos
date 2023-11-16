package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rumenvasilev/go-fritzos/auth"
	"github.com/rumenvasilev/go-fritzos/nas"
)

func main() {
	username := ""
	password := ""

	Session, err := auth.Auth(username, password)
	defer auth.Close(Session)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Login successful! Session ID", Session)

	// List all files at the root
	// res, err := ListDirectory(Session)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// List files at specific directory level
	// params := make(map[string]string)
	// params["path"] = "/Bilder"
	// params["limit"] = "100"
	// res, err := ListDirectoryWithParams(Session, params)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// dirs := res.Directories
	// files := res.Files

	// if len(dirs) > 0 {
	// 	fmt.Println("Listing directories:")
	// }
	// for _, v := range dirs {
	// 	fmt.Println(v.Filename)
	// }

	// if len(files) > 0 {
	// 	fmt.Println("Listing files:")
	// }
	// for _, v := range files {
	// 	fmt.Println(v.Filename)
	// 	fmt.Println(v.Size)
	// 	fmt.Println(v.Path)
	// 	fmt.Println(v.Timestamp)
	// }

	// Get specific object from the NAS
	// d, err := GetFile(Session, "/Bilder/FRITZ-Picture.jpg")
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// data, _ := io.ReadAll(d)
	// defer d.Close()

	// err = os.WriteFile("FRITZ-Picture.jpg", data, 0755)
	// if err != nil {
	// 	log.Fatalln("failed writing the resulting file to disk, %w", err)
	// }

	// Put File
	// data, err := os.Open("/Users/rumba/Downloads/2022-0006_Anmeldung_fr_den_Schwimmkurs_Einverstndniserklrung.pdf")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// r, err := PutFile(Session, "/Dokumente/2022-0006_Anmeldung_fr_den_Schwimmkurs_Einverstndniserklrung.pdf", data)
	// if err != nil {
	// 	log.Fatalln("failed uploading the file, %w", err)
	// }
	// if r.ResultCode != nasResultOK || r.SuccessfulUploads == nasUploadFail {
	// 	log.Fatalln("Upload failed")
	// }
	// s, _ := json.Marshal(r)
	// fmt.Println(string(s))

	// Rename File
	// r, err := RenameFile(Session, "/Dokumente/blabla.pdf", "2022-0006-baba.pdf")
	// if err != nil {
	// 	log.Fatalf("Rename failed, %v", err)
	// }
	// s, _ := json.Marshal(r)
	// fmt.Println(string(s))

	// Delete File
	// r, err := DeleteFile(Session, "/Dokumente/2022-0006-baba.pdf")
	// if err != nil {
	// 	log.Fatalf("Rename failed, %v", err)
	// }
	// s, _ := json.Marshal(r)
	// fmt.Println(string(s))

	// Move File
	// n, err := MoveFile(Session, "/2022-0006_Anmeldung_fr_den_Schwimmkurs_Einverstndniserklrung.pdf", "/Dokumente")
	// if err != nil {
	// 	log.Fatalf("Move failed, %v", err)
	// }
	// fmt.Println(n)

	// Create Dir
	n := nas.New(Session).WithAddress("http://fritz.box")

	r, err := n.CreateDir("blabla", "/Dokumente")
	if err != nil {
		log.Fatalf("Create failed, %v", err)
	}
	s, _ := json.Marshal(r)
	fmt.Println(string(s))

	log.Println("Logged out.")
}
