package main

import(
	"os"
	"fmt"
	"log"
	"io"
	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"golang.org/x/net/context"
	"strings"
)

func main(){
	os.Setenv("GOOGLE_CLOUD_PROJECT",os.Args[1])
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS",os.Args[2])
	sql := os.Args[3]
	outfile := os.Args[4]

	proj := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if proj == "" {
		fmt.Println("GOOGLE_CLOUD_PROJECT environment variable must be set.")
		os.Exit(1)
	}

	rows,err := query(proj,sql)
	if err != nil {
		log.Fatal(err)
	}
	if err := saveResults(rows, outfile); err != nil {
		log.Fatal(err)
	}
}

//query runs a sql like statement provided in the arguments
//Returns a bigquery rowiterator
func query(proj, sql string)(*bigquery.RowIterator, error){
	ctx := context.Background()

	client,err := bigquery.NewClient(ctx,proj)
	if err != nil {
		return nil, err
	}

	query := client.Query(sql)
	return query.Read(ctx)
}

//saves the results of the query based on the extension provided in the command line arguments
//currently only json, csv and tsv are supported
func saveResults(iter *bigquery.RowIterator, outfile string) error {

	ext := strings.Split(outfile,".")[1]

	file, err := os.Create(outfile)
	if err != nil {
		log.Fatal("Cannot create file",err)
	}
	defer file.Close()

	switch ext {
	case "json":
		err = saveJson(file,iter)
	case "csv":
		err = saveDelim(file,iter,",")
	case "tsv":
		err = saveDelim(file,iter,"\t")
	default:
		os.Remove(outfile)
		fmt.Println(ext + " is not a currently supported output type.")
		os.Exit(1)
	}

	if err != nil{
		return err
	}
	return nil
}

//function to save to JSON
//as a list of dictionaries
//each row a dictionary, key is the column name
func saveJson(file io.Writer, iter *bigquery.RowIterator) error {
	rowstr := "[\n"
	for{
		var row map[string]bigquery.Value
		err := iter.Next(&row)
		if err == iterator.Done {
			rowstr = rowstr[:len(rowstr)-2]
			rowstr += "]\n"
			fmt.Fprintf(file,rowstr)
			return nil
		}
		if err != nil {
			return err
		}

		rowstr += "{"

		for k,v := range row {
			if _,ok := v.(string); ok != false{
				rowstr += fmt.Sprintf("\"%s\": \"%v\",",k,v)
			} else {
			rowstr += fmt.Sprintf("\"%s\": %v,",k,v)
		}
		}
		rowstr = rowstr[:len(rowstr)-1]
		rowstr += "},\n"
	}

	return nil
}

//function to save delimited file
//currently there no headers, just rows of observations
func saveDelim(file io.Writer, iter *bigquery.RowIterator, delim string) error {
	for{
		var row []bigquery.Value
		err := iter.Next(&row)
		if err == iterator.Done{
			return nil
		}
		if err != nil{
			return err
		}
		for _, v := range row {
			if row[len(row)-1] == v {
				fmt.Fprintf(file,"\"%v\"\n",v)
			} else {
				fmt.Fprintf(file,"\"%v\"" + delim,v)
			}
		}
	}
	return nil
}
