package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	username := ""
	password := ""

	sessionID, err := Auth(username, password)
	defer Close(sessionID)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Login successful! Session ID", sessionID)

	// List all files at the root
	// res, err := ListDirectory(sessionID)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// List files at specific directory level
	// params := make(map[string]string)
	// params["path"] = "/Bilder"
	// params["limit"] = "100"
	// res, err := ListDirectoryWithParams(sessionID, params)
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
	// d, err := GetFile(sessionID, "/Bilder/FRITZ-Picture.jpg")
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
	// r, err := PutFile(sessionID, "/Dokumente/2022-0006_Anmeldung_fr_den_Schwimmkurs_Einverstndniserklrung.pdf", data)
	// if err != nil {
	// 	log.Fatalln("failed uploading the file, %w", err)
	// }
	// if r.ResultCode != nasResultOK || r.SuccessfulUploads == nasUploadFail {
	// 	log.Fatalln("Upload failed")
	// }
	// s, _ := json.Marshal(r)
	// fmt.Println(string(s))

	// Rename File
	// r, err := RenameFile(sessionID, "/Dokumente/blabla.pdf", "2022-0006-baba.pdf")
	// if err != nil {
	// 	log.Fatalf("Rename failed, %v", err)
	// }
	// s, _ := json.Marshal(r)
	// fmt.Println(string(s))

	// Delete File
	// r, err := DeleteFile(sessionID, "/Dokumente/2022-0006-baba.pdf")
	// if err != nil {
	// 	log.Fatalf("Rename failed, %v", err)
	// }
	// s, _ := json.Marshal(r)
	// fmt.Println(string(s))

	// Move File
	// n, err := MoveFile(sessionID, "/2022-0006_Anmeldung_fr_den_Schwimmkurs_Einverstndniserklrung.pdf", "/Dokumente")
	// if err != nil {
	// 	log.Fatalf("Move failed, %v", err)
	// }
	// fmt.Println(n)

	// Create Dir
	r, err := CreateDir(sessionID, "blabla", "/Dokumente")
	if err != nil {
		log.Fatalf("Create failed, %v", err)
	}
	s, _ := json.Marshal(r)
	fmt.Println(string(s))

	log.Println("Logged out.")
}