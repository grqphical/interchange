package templates

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

const dirTemplate string = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ printf "Contents of %s/" .Directory }}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
            background-color: #f9f9f9;
            color: #333;
        }

        .directory-header {
            text-align: center;
            margin: 20px 0;
            font-size: 24px;
            color: #444;
        }

        .directory-table {
            width: 80%;
            margin: 0 auto;
            border-collapse: collapse;
            background-color: #fff;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        .directory-table th, .directory-table td {
            padding: 10px;
            text-align: left;
            border: 1px solid #ddd;
        }

        .directory-table th {
            background-color: #f4f4f4;
            font-weight: bold;
        }

        .directory-table tr:nth-child(even) {
            background-color: #f9f9f9;
        }

        .directory-table tbody tr:hover {
            background-color: #f1f1f1;
            cursor: pointer;
        }
    </style>
</head>
<body>
    <h1 class="directory-header">{{ printf "Contents of %s/" .Directory }}</h1>
    {{if not .IsRoot}}
        <div style="text-align: center; margin: 20px;">
            <a href="../" style="text-decoration: none; color: #fff; background-color: #007BFF; padding: 10px 20px; border-radius: 5px; font-size: 16px;">Go to Parent Directory</a>
        </div>
    {{end}}
    <table class="directory-table">
        <thead>
            <tr>
                <th>Name</th>
                <th>Size</th>
                <th>Last Modified</th>
            </tr>
        </thead>
        <tbody>
		{{range $file := .Files}}
            <tr>
                <td><a href={{$file.FileURL}}>{{$file.Name}}</a></td>
                <td>{{$file.Size}}</td>
                <td>{{$file.Date}}</td>
            </tr>
		{{end}}
        </tbody>
    </table>
    <footer style="text-align: center; margin: 20px 0; font-size: 14px; color: #666;">
        {{ .Version }}
    </footer>
</body>
</html>`

// information about individual files to be used in the template
type fileInfo struct {
	Name    string
	FileURL string
	Size    string
	Date    string
}

// information about the current directory to be used in the template
type directoryParams struct {
	Files     []fileInfo
	Directory string
	Version   string
	IsRoot    bool
}

// takes in an integer file size (in bytes) and returns a formatted string that included the greatest size unit.
// Example 1024 -> 1KB
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// write the directory browser template to the given http.ResponseWriter
func WriteDirectoryTemplate(w http.ResponseWriter, dir string, baseURL string, baseDirectory string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	var dirTitleString string
	if dir == baseDirectory {
		dirTitleString = filepath.Base(dir)
	} else {
		dirTitleString = filepath.Join(filepath.Base(baseDirectory), filepath.Base(dir))
	}

	params := directoryParams{
		Files:     make([]fileInfo, len(files)),
		Directory: dirTitleString,
		Version:   serverString,
		IsRoot:    baseDirectory == dir,
	}

	for i := 0; i < len(files); i++ {
		info, err := files[i].Info()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		params.Files[i] = fileInfo{
			Name:    files[i].Name(),
			FileURL: path.Join(baseURL, files[i].Name()),
			Size:    formatFileSize(info.Size()),
			Date:    info.ModTime().Format(time.DateTime),
		}
	}

	tmpl := template.Must(template.New("directory").Parse(dirTemplate))
	tmpl.Execute(w, params)
}
