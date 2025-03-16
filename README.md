# **mega-stream: A Go Library for MEGA File Streaming & Decryption**

`mega-stream` is a powerful Go library designed for **streaming and decrypting** MEGA files. It handles downloading encrypted chunks from MEGA servers, decrypting them, and allows for streaming of large files (e.g., videos or other media) in a controlled and efficient way. This library can be easily imported into any Go application to handle MEGA links and stream them as needed.

---

## **Table of Contents**

- [Features](#features)
- [Project Structure](#project-structure)
- [Installation](#installation)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

---

## **Features**

- 🔐 **AES Decryption**: Handles AES decryption of MEGA files using proper key and IV management.
- 🌐 **Streaming Support**: Supports partial content downloads and streaming from MEGA servers.
- 🧩 **Modular Architecture**: Built with a clean, modular, and reusable codebase for easy integration into Go projects.
- 🚀 **Efficient Chunk Downloading**: Downloads file chunks on-demand, reducing memory usage and optimizing bandwidth.

---

## **Project Structure**

```plaintext
mega-stream/
│
├── go.mod
├── go.sum
│
├── cmd/
│   └── demo-server/              # Optional: CLI/demo server to test the lib
│       └── main.go
│
├── internal/                     # Internal packages (not meant to be imported outside)
│   ├── crypto/
│   │   └── decrypt.go            # AES decryption, IV handling, key parsing
│   ├── megaapi/
│   │   ├── downloader.go        # HTTP chunk downloader
│   │   └── metadata.go          # MEGA file metadata retrieval
│   └── utils/
│       └── parser.go            # URL parser, helpers
│
├── pkg/
│   └── megastream/
│       ├── stream.go            # Public package entry: streaming logic
│       ├── file.go              # MEGA file abstraction (size, mimeType, etc.)
│       └── types.go             # Common types and interfaces
│
└── test/
    ├── testdata/
    │   └── sample.megalink.txt  # Test links
    └── stream_test.go           # Unit/integration tests
```

### **Key Folders and Files:**
- **`pkg/megastream/`**: Contains the public API for streaming MEGA files, including methods like `NewFromURL()` and `StreamChunk()`.
- **`internal/`**: Internal packages that manage encryption, chunk downloading, and metadata retrieval from MEGA.
- **`cmd/demo-server/`**: Optional demo server for testing the library.
- **`test/`**: Unit tests and sample MEGA links used for integration testing.

---

## **Installation**

To install `mega-stream` in your Go project, use the following command:

```bash
go get github.com/valle-tech/mega-streaming
```

Ensure that you also have Go modules enabled (`go mod init`), as this library is structured to work with Go modules.

---

## **Usage**

Here’s how to use the `mega-stream` library in your own Go application to stream MEGA files.

1. **Import the library**:

   ```go
   import (
       "github.com/valle-tech/mega-streaming/pkg/megastream"
   )
   ```

2. **Handle Streaming**:

   ```go
   func handleVideoStream(w http.ResponseWriter, r *http.Request) {
       url := r.URL.Query().Get("encodedUrl")

       file, err := megastream.NewFromURL(url)
       if err != nil {
           http.Error(w, "Invalid MEGA URL", http.StatusBadRequest)
           return
       }

       start, end := parseRange(r.Header.Get("Range"), file.Size)

       streamReader, err := file.StreamChunk(start, end)
       if err != nil {
           http.Error(w, "Failed to stream file", http.StatusInternalServerError)
           return
       }

       w.Header().Set("Content-Type", file.MimeType)
       w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.Size))
       w.Header().Set("Content-Length", strconv.Itoa(int(end-start+1)))
       w.WriteHeader(http.StatusPartialContent)

       io.Copy(w, streamReader)
   }
   ```

3. **Stream a file**:

   You can call the `NewFromURL()` function to create a `MegaFile` from a MEGA URL. Then, call `StreamChunk()` to stream chunks from the file.

---

## **API Reference**

### **MegaFile Struct**

```go
type MegaFile struct {
    Size     int64  // Total size of the file
    MimeType string // MIME type of the file (e.g., video/mp4)
}
```

### **NewFromURL**

```go
func NewFromURL(encodedUrl string) (*MegaFile, error)
```

**Parameters**:  
- `encodedUrl` (string): The MEGA URL.

**Returns**:  
- A `MegaFile` object containing file size and MIME type.
- An error if the URL is invalid or if decryption fails.

### **StreamChunk**

```go
func (f *MegaFile) StreamChunk(start, end int64) (io.Reader, error)
```

**Parameters**:  
- `start` (int64): The starting byte range for streaming.
- `end` (int64): The ending byte range for streaming.

**Returns**:  
- An `io.Reader` for streaming the file chunk.
- An error if the chunk cannot be retrieved or decrypted.

---

## **Testing**

### Unit Tests

To run unit tests for the library:

```bash
go test ./...
```

Tests are located in the `test/` directory, and you can write your own test cases as needed.

### Running Demo Server

A simple demo server is included to test the library in action. To start the server:

```bash
go run cmd/demo-server/main.go
```

---

## **Contributing**

We welcome contributions! Please follow these steps to contribute:

1. **Fork the repository** and clone it locally.
2. **Create a feature branch**: `git checkout -b feature/your-feature-name`.
3. **Make your changes** and ensure the code is properly tested.
4. **Commit your changes**: `git commit -m 'Add your feature'`.
5. **Push to your fork**: `git push origin feature/your-feature-name`.
6. **Create a pull request** to the main repository.

### **Code Style**
- Follow Go's idiomatic style.
- Use `gofmt` for formatting.
- Provide clear commit messages.

### **Issues**
If you find a bug or have a feature request, feel free to open an issue on the repository.

---

## **License**

`mega-stream` is open-source and available under the **MIT License**.

---
