package example

import (
    "bytes"
    "context"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "github.com/elastic/go-elasticsearch/v7"
    "github.com/elastic/go-elasticsearch/v7/esapi"
    "github.com/elastic/go-elasticsearch/v7/esutil"
    "log"
    "net"
    "net/http"
    "strings"
    "time"
)

var (
    es *elasticsearch.Client
    err error
    indexName = "test-000001"
)

type user struct {
    ID	int			`json:"userid"`
    Title string	`json:"title"`
    Tags []string	`json:"tags,omitempty"`
}

func main(){
    Conn()

    //Info()
    //log.Println(strings.Repeat("▔", 65))
    //
    ////DeleteIndex()
    ////log.Println(strings.Repeat("▔", 65))
    //CreateIndex()
    //log.Println(strings.Repeat("▔", 65))
    //
    //InsertOne()
    //log.Println(strings.Repeat("▔", 65))
    //
    //InserMany()

    //Update()

    Translate()
}

func Conn(){
    cfg := elasticsearch.Config{
        Addresses: []string{
            "http://localhost:9200",
        },
        //Username: "userName",
        //Password: "pwd",
        Transport: &http.Transport{
            MaxIdleConnsPerHost:   10,
            ResponseHeaderTimeout: time.Second,
            DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
            TLSClientConfig: &tls.Config{
                MinVersion:         tls.VersionTLS11,
            },
        },
    }
    es, err = elasticsearch.NewClient(cfg)
}

func Info(){
    res, err := es.Info()
    if err != nil {
        log.Fatalf("Error getting response: %s", err)
    }
    defer res.Body.Close()
    fmt.Println(res)
}

func CreateIndex(){
    mapping := `{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"properties":{
			"userid":{
				"type":"keyword"
			},
			"title":{
				"type":"text",
				"store": true,
				"fielddata": true,
				"analyzer": "ik_max_word",
            	"search_analyzer": "ik_smart"
			},
			"tags":{
				"type":"keyword"
			}
		}
	}
}`
    req := esapi.IndicesCreateRequest{
        Index: indexName,
        Body:  strings.NewReader(mapping),
    }
    res, err := req.Do(context.Background(), es)
    if err != nil {
        fmt.Println("创建索引错误：", err)
    }
    defer res.Body.Close()
    fmt.Println(res)
}

func DeleteIndex() {
    req := esapi.IndicesDeleteRequest{
        Index: []string{indexName},
    }
    res, err := req.Do(context.Background(), es)
    if err !=nil {
        fmt.Println("删除索引错误：", err)
    }
    defer res.Body.Close()
    fmt.Println(res)
}


func InsertOne() {
    user1 := user{1, "PHP工程师", []string{"男", "黄种人", "PHP"}}
    jsonBody, _ := json.Marshal(user1)
    req := esapi.CreateRequest{    // 如果是esapi.IndexRequest则是插入/替换
        Index:        indexName,
        DocumentType: "_doc",
        DocumentID:   fmt.Sprint(user1.ID),
        Body:         bytes.NewReader(jsonBody),
    }
    res, err := req.Do(context.Background(), es)
    if err !=nil {
        fmt.Println("插入一条记录错误：", err)
    }
    defer res.Body.Close()
    fmt.Println(res)
}

func InserMany(){
    bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
        Index:         indexName,        // The default index name
        Client:        es,               // The Elasticsearch client
        NumWorkers:    2,       // The number of worker goroutines
        FlushBytes:    5e+6,  // The flush threshold in bytes
        FlushInterval: 30 * time.Second, // The periodic flush interval
    })
    start := time.Now().UTC()
    ctx := context.Background()
    for i:=2; i<12; i++ {
        u := user{
            i,
            fmt.Sprintf("PHP工程师-%d", i),
            []string{"男", fmt.Sprint(i)},
        }
        jsonBody, _ := json.Marshal(u)
        bi.Add(ctx,
            esutil.BulkIndexerItem{
                // Action field configures the operation to perform (index, create, delete, update)
                Action: "index",

                // DocumentID is the (optional) document ID
                DocumentID: fmt.Sprint(u.ID),

                // Body is an `io.Reader` with the payload
                Body: bytes.NewReader(jsonBody),

                // OnSuccess is called for each successful operation
                OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
                    //atomic.AddUint64(&countSuccessful, 1)
                },

                // OnFailure is called for each failed operation
                OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
                    if err != nil {
                        log.Printf("ERROR: %s", err)
                    } else {
                        log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
                    }
                },
            })

    }


    // Close the indexer
    //
    if err := bi.Close(context.Background()); err != nil {
        log.Fatalf("Unexpected error: %s", err)
    }

    biStats := bi.Stats()

    // Report the results: number of indexed docs, number of errors, duration, indexing rate
    //
    log.Println(strings.Repeat("▔", 65))

    dur := time.Since(start)

    if biStats.NumFailed > 0 {
        log.Fatalf(
            "Indexed [%d] documents with [%s] errors in %s (%s docs/sec)",
            int64(biStats.NumFlushed),
            biStats.NumFailed,
            dur.Truncate(time.Millisecond),
            int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed)),
        )
    } else {
        log.Printf(
            "Sucessfuly indexed [%d] documents in %s (%d docs/sec)",
            int64(biStats.NumFlushed),
            dur.Truncate(time.Millisecond),
            int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed)),
        )
    }
}

func Update(){
    doc := `{
	"query": {
	"bool": {
	  "must": [
		{"match_phrase_prefix": {
		  "title": {
			"query": "-5"
		  }
		}}
	  ]
	}
	},
	"script":{
		"source": "ctx._source.title=params.title;",
		"params": {
			"title": "公司水电费水电费aaaaaaaa"
		}
	}

}
`
    req := esapi.UpdateByQueryRequest{
        Index: []string{indexName},
        Body:  strings.NewReader(doc),
    }
    res, err := req.Do(context.Background(), es)

    defer res.Body.Close()
    fmt.Println(res.String())
    if err != nil {
        log.Fatal(err, "Error Update response")
    }
    fmt.Println(res.String())
}

func Translate(){
    //    sql := `{
    //  "query": "select title, userid  from \"test-000001\" where title like '%php%'  limit 2"
    //}
    //`
    q := map[string]interface{}{
        "query": `select title, userid  from "test-000001"   limit 2`,
    }
    b, _ := json.Marshal(q)

    req := esapi.SQLQueryRequest{
        Format: "json",
        Body: bytes.NewReader(b),
        //Body: strings.NewReader(sql),
    }
    res, err := req.Do(context.Background(), es)
    if err != nil {
        log.Fatal(err, "Error Update response")
    }
    fmt.Println(res.String())
}
