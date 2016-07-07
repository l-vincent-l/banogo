package main
import (
    "github.com/kataras/iris"
        "os"
       "fmt"
      . "github.com/l-vincent-l/trigram"
       "encoding/json"
       "compress/gzip"
        "io"
        "log"
        "strings"
        _ "net/http/pprof"
        "net/http"
)

var ti *TrigramIndex

type BanoLine struct {
    Name string `json: name`
    Postcode string `json: postcode`
    City string `json: city`
    Departement string `json: department`
    Region string `json: region`
    Type string `json: type`
}

var docs []BanoLine
func readfile(path string, ti *TrigramIndex) {
    inFile, _ := os.Open(path)
    log.Println(path)
    defer inFile.Close()
    dec := json.NewDecoder(inFile)
    for {
        var bl BanoLine
        if err := dec.Decode(&bl); err == io.EOF {
            break
        } else if err != nil {
            log.Println(err)
            log.Println(len(docs))
            break
            continue
        }
        if bl.Type != "street" {
            continue
        }

        doc := fmt.Sprintf("%s %s %s %s %s", bl.Name, bl.City, bl.Postcode,
                    bl.Departement, bl.Region)
        docs = append(docs, bl)
        ti.Add(strings.ToLower(doc))
        if len(docs)%1000 ==0 {
          log.Println("%d lines read", len(docs))
        }
    }
    fmt.Println(fmt.Sprintf("map len: %d", len(ti.TrigramMap)))
}

func search(c *iris.Context) {
    q := c.URLParam("q")
    var m []BanoLine
    for _, value := range ti.Query(q) {
        m = append(m, docs[value-1])
    }
    c.JSON(iris.StatusOK, m)
}

func download_departement(departement int) string {
    url := fmt.Sprintf("https://bano.openstreetmap.fr/data/bano-%02d.json.gz", departement)
    tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	fmt.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return ""
	}
	defer output.Close()

    client := new(http.Client)
    request, err := http.NewRequest("GET", url, nil)
    request.Header.Add("Accept-Encoding", "gzip")

    response, err := client.Do(request)
    defer response.Body.Close()

    // Check that the server actually sent compressed data
    var reader io.ReadCloser
    reader, err = gzip.NewReader(response.Body)
    defer reader.Close()

	n, err := io.Copy(output, reader)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
        return ""
	}

    fmt.Println(n, "bytes downloaded.")
    return fileName
}

func initIndex(ti *TrigramIndex) {
    for i:=1; i <=95; i++ {
        fileName := download_departement(i)
        if fileName != "" {
            readfile(fileName, ti)
        }
    }
}

func main() {
    ti = NewTrigramIndex()
    initIndex(ti)
    iris.Get("/search", search)
    iris.Listen(":8080")
}
