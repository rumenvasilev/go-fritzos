# go-fritzos
Golang SDK for Fritz! OS devices

Still very unstable and under active development. API will most probably change.

Supported operations:

* Authentication v2 (pbkdf), v1 (md5)
* NAS - `get`, `put`, `delete`, `move`, `rename`, `list`, `createdir` for both files and directories

Examples:

See the [example](example) directory for CLI implementation with all currently supported features.

Testing/development platform:
* Fritz!OS 7.57, running on Fritz!Box 6591 Cable.
