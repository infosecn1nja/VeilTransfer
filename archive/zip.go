package archive

import (
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"
)

type FileSize struct {
    Size int64
    Path string
}

type ByLargest []FileSize

func (f ByLargest) Len() int           { return len(f) }
func (f ByLargest) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByLargest) Less(i, j int) bool { return f[i].Size > f[j].Size }

func loadFiles(path string) ([]FileSize, error) {
    var ret []FileSize
    err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            ret = append(ret, FileSize{
                Size: 0,
                Path: filePath,
            })
        } else {
            ret = append(ret, FileSize{
                Size: info.Size(),
                Path: filePath,
            })
        }
        return nil
    })
    return ret, err
}

func CreateZip(localPath, outputPath string, splitSize int64) error {
    if splitSize == 0 {
        // No split size provided, create a single zip file
        zipfile, err := os.Create(outputPath)
        if err != nil {
            return fmt.Errorf("failed to create zip file: %w", err)
        }
        defer zipfile.Close()

        zw := zip.NewWriter(zipfile)
        defer zw.Close()

        return addFilesToZip(zw, localPath, localPath)
    }

    // Split size provided, create split zip files
    files, err := loadFiles(localPath)
    if err != nil {
        return err
    }

    groupedFiles, err := groupBySize(files, splitSize, 0)
    if err != nil {
        return err
    }

    prefix := strings.TrimSuffix(filepath.Base(outputPath), filepath.Ext(outputPath))
    baseDir := localPath
    return splitZip(groupedFiles, filepath.Dir(outputPath), prefix, baseDir)
}

func addFilesToZip(zw *zip.Writer, baseDir, localPath string) error {
    return filepath.Walk(localPath, func(filePath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            return nil
        }

        relPath, err := filepath.Rel(baseDir, filePath)
        if err != nil {
            return err
        }
        relPath = filepath.ToSlash(relPath)

        header, err := zip.FileInfoHeader(info)
        if err != nil {
            return err
        }
        header.Name = relPath
        header.Method = zip.Deflate

        writer, err := zw.CreateHeader(header)
        if err != nil {
            return err
        }

        file, err := os.Open(filePath)
        if err != nil {
            return err
        }
        defer file.Close()

        _, err = io.Copy(writer, file)
        return err
    })
}

func splitZip(groupedFiles [][]string, outpath string, prefix string, baseDir string) error {
    outpath, _ = filepath.Abs(outpath)

    for i, g := range groupedFiles {
        zipfileName := fmt.Sprintf("%s.%03d.zip", prefix, i+1)
        zipfile, err := os.Create(filepath.Join(outpath, zipfileName))
        if err != nil {
            return err
        }
        zw := zip.NewWriter(zipfile)

        for _, f := range g {
            info, err := os.Stat(f)
            if err != nil {
                return err
            }
            isDir := info.IsDir()
            if err := addFileToZip(zw, f, baseDir, isDir); err != nil {
                return err
            }
        }

        if err := zw.Close(); err != nil {
            return err
        }
        if err := zipfile.Close(); err != nil {
            return err
        }
    }
    return nil
}

func addFileToZip(zw *zip.Writer, filePath, baseDir string, isDir bool) error {
    relativePath, err := filepath.Rel(baseDir, filePath)
    if err != nil {
        return err
    }
    relativePath = filepath.ToSlash(relativePath)

    if isDir {
        relativePath += "/"
        header := &zip.FileHeader{
            Name:     relativePath,
            Method:   zip.Store,
            Modified: time.Now(),
        }
        _, err := zw.CreateHeader(header)
        return err
    }

    zf, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer zf.Close()

    info, err := zf.Stat()
    if err != nil {
        return err
    }

    header, err := zip.FileInfoHeader(info)
    if err != nil {
        return err
    }

    header.Method = zip.Deflate
    header.Name = relativePath

    writer, err := zw.CreateHeader(header)
    if err != nil {
        return err
    }

    _, err = io.Copy(writer, zf)
    return err
}

func groupBySize(files []FileSize, maxSize int64, savings float64) ([][]string, error) {
    var ret [][]string
    maxSize = maxSize + int64(float64(maxSize)*savings)

    sort.Sort(ByLargest(files))
    for i, e := range files {
        if e.Size == -1 {
            continue
        }

        if e.Size > maxSize {
            fmt.Printf("[!] Warning: File %s exceeds the split size, putting it in its own zip part.\n", e.Path)
            ret = append(ret, []string{e.Path})
            files[i].Size = -1
            continue
        }

        files[i].Size = -1
        group := []string{e.Path}
        d := maxSize - e.Size

        j := i + 1
        s := e.Size
        for ; j < len(files); j++ {
            if files[j].Size == -1 {
                continue
            }
            if files[j].Size <= d {
                s += files[j].Size
                d = maxSize - s
                group = append(group, files[j].Path)
                files[j].Size = -1
            }
            if s > maxSize {
                break
            }
        }
        ret = append(ret, group)
    }
    return ret, nil
}