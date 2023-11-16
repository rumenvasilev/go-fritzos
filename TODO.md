# TODO

* All requests use context and have timeout set
* RenameFile:

```
Rename failed, return status code 400, response:
"{\"error\":{\"message\":\"[file_system_controller] An error occurred while renaming files or folders.\",\"data\":[{\"message\":\"Because one of the parameters submitted was incorrect, renaming cannot be carried out.\",\"path\":\"\\/Dokumente\\/2022-0006-baba.pdf\",\"code\":9}],\"code\":400}}"
```

```
{\"error\":{\"message\":\"[file_system_controller] The folder with the name \\\"blabla\\\" already exists and therefore cannot be created.\",\"data\":{\"message\":\"The folder with the name \\\"blabla\\\" already exists and therefore cannot be created.\",\"path\":\"\\/Dokumente\",\"code\":5},\"code\":400}}
```
