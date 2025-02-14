package handlers

import (
	"fmt"
	"net/http"
	"os"
	"golang.org/x/net/webdav"
)

const verbose = true

func basicAuth(next http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != username || p != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			if verbose {
				fmt.Printf("Unauthorized access attempt from %s\n", r.RemoteAddr)
			}
			return
		}
		if verbose {
			fmt.Printf("Authorized access by user '%s' from %s\n", u, r.RemoteAddr)
		}
		next.ServeHTTP(w, r)
	}
}

func StartWebdav(useSSL bool, certFile, keyFile, port, destDir, username, password string) {
	if username == "" || password == "" {
		fmt.Println("Username and password are required for authentication")
		os.Exit(1)
	}

	handler := &webdav.Handler{
		Prefix:     "/",
		FileSystem: webdav.Dir(destDir),
		LockSystem: webdav.NewMemLS(),
	}

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err = os.Mkdir(destDir, 0755)
		if err != nil {
			fmt.Printf("Failed to create destination directory: %v\n", err)
			os.Exit(1)
		}
		if verbose {
			fmt.Printf("Created destination directory: %s\n", destDir)
		}
	} else if verbose {
		fmt.Printf("Using existing destination directory: %s\n", destDir)
	}

	http.HandleFunc("/", basicAuth(handler.ServeHTTP, username, password))

	addr := ":" + port

	if useSSL {
		fmt.Printf("Starting WebDAV server on %s (HTTPS)\n", addr)
		if verbose {
			fmt.Printf("SSL enabled with certificate: %s, key: %s\n", certFile, keyFile)
		}
		if err := http.ListenAndServeTLS(addr, certFile, keyFile, nil); err != nil {
			fmt.Printf("Failed to start HTTPS server: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Starting WebDAV server on %s (HTTP)\n", addr)
		if verbose {
			fmt.Println("SSL is not enabled. Running in HTTP mode.")
		}
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Printf("Failed to start HTTP server: %v\n", err)
			os.Exit(1)
		}
	}
}
