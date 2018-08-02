package main

import(
	"os"
	"os/exec"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"net/rpc"

	qt "querytool"
	ps "github.com/mitchellh/go-ps"
)

// feeder executable starts with a single argument to the feeder configuration file
// saves logs to "querytool_logs.txt", in the same folder as the executable
// feeder then validates if the configuraiton file exists before marshalling into a map
// feeder than checks if the receiver executable is running and starts it if it isn't
// feeder then identifies all .src file in the "root_jobs_dir" and reads these into source objects
// performs the same process for .conf files
// combines information from source and config objects into a job object which is then sent over to the receiver
func main(){

	logfile, err := os.OpenFile("querytool_logs.txt",os.O_CREATE|os.O_APPEND|os.O_WRONLY,0644)
	if err != nil {
		log.Fatal(err)
	}

	defer logfile.Close()
	log.SetOutput(logfile)

	if len(os.Args) < 2 {
		log.Fatal("Missing first argument. Please include the path to feed_config.json.")
	}

	feedconfig :=  readFeedConfig(os.Args[1])

	err = checkRetriever()
	if err != nil{
		log.Fatal(err)
	}

	sourcefiles := identifyFiles(feedconfig["root_jobs_dir"],".src")
	sourceobjects := processSources(sourcefiles,feedconfig["parse_validate_map"])

	configfiles := identifyFiles(feedconfig["root_jobs_dir"],".conf")
	jobs := processConfigs(configfiles,sourceobjects)

	sendToReceiver(jobs)
}


// Reads the feeder configuration json file from the first command line argument
// If file is not found or cannot be read throws an error
// If file cannot be marshalled from json into a map an error is thrown
// Returns a map of strings with a directory name and the directory path
func readFeedConfig(filepath string) (map[string]string){

	configmap, err := qt.JsonToMap(filepath)
	if err != nil {
		log.Fatal(err)
	}

	return configmap
}

// Checks if the retriever executable is running
// Starts the retriever if it is not already running
// Detaches the process once started
func checkRetriever() error{
	processes, err := ps.Processes()

	if err != nil{
		return err
	}
	for _,p := range processes {
		if p.Executable() == "receiver.exe" {
			return nil
		}
	}

	//include /b after start to close the command line GUI
	cmnd := exec.Command("cmd.exe","/C","start","../receiver/receiver.exe")
	if err := cmnd.Run(); err != nil {
		return err
	}

	return nil
}

// Checks for all files within the directory and subdirectories with the provided extension
// Returns a slice with all files with the extension
func identifyFiles(rootfiledir, ext string) []string {
	var files []string

	err := filepath.Walk(rootfiledir, func(path string, info os.FileInfo, err error) error {
		files = append(files,path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	var filelist []string

	for _, file := range files {
		if fileext := filepath.Ext(file); fileext == ext {
			filelist = append(filelist,file)
		}
	}

	return filelist
}

// takes a slice of source files and reads the contents
// unmarshals the contents to a source object
// panics if there are any issues marshaling
// logs if there are any changes to the source file
func processSources(sourcefiles []string, hashtable string) []qt.Source {
	var sources []qt.Source

	for _, file := range sourcefiles {

		msg, err := qt.HasFileChanged(hashtable,file)
		if err != nil{
			log.Fatal(err)
		}
		if msg != ""{
			log.Println(msg)
		}

		contents, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		var s qt.Source
		if err := json.Unmarshal(contents, &s); err != nil {
			log.Println(err)
		}

		//only requires name, type and connection string for now
		if s.Name == "" || s.Type == "" || s.ConnStr == "" {
			fmt.Printf("Ignoring source file \"%s\". All source files must have the following fields: \"name\", \"type\", \"connection_string\"\n",file)
			continue
		}

		sources = append(sources,s)
	}

	return sources
}

// Move through configuration files list and check cron schedule to see if it should be run
// Unmarshall into go object based on the type
// Validate the object by checking for zero values
// Combine with the inbound and outbound sources and marshall into a JSON to be sent to the receiver
func processConfigs(configfiles []string, sources []qt.Source) []qt.Job {
	var jobs []qt.Job

	for _,config := range configfiles {
		contents, err := ioutil.ReadFile(config)
		if err != nil{
			log.Fatal(err)
		}

		var configmap map[string]*json.RawMessage
		err = json.Unmarshal(contents, &configmap)
		if err != nil{
			log.Println(err)
			continue
		}

		// Note: The "type" will be a broad grouping for many types of data sources
		// The "type" will need to be included in the .conf json file, along with the specified fields for the specific configuration type object
		var configtype string
		err = json.Unmarshal(*configmap["type"],&configtype)
		if err != nil{
			log.Println(err)
			continue
		}

		job := qt.Job{}
		switch configtype {
			case "RDB":
				var RDBobj qt.RDBConfig
				if err := json.Unmarshal(contents, &RDBobj); err != nil {
					log.Println(err)
					continue
				}

				for i := range sources {
					if sources[i].Name == RDBobj.Source {
						job.Source = sources[i]
					}
					if sources[i].Name == RDBobj.Destination {
						job.Destination = sources[i]
					}
				}

				job.Queries = RDBobj.Queries
				job.IsDist = false
			case "DFS":
				var DFSobj qt.DFSConfig
				if err := json.Unmarshal(contents, &DFSobj); err != nil {
					log.Println(err)
					continue
				}

				for i := range sources {
					if sources[i].Name == DFSobj.Source {
						job.Source = sources[i]
					}
					if sources[i].Name == DFSobj.Destination {
						job.Destination = sources[i]
					}
				}

				job.Queries = DFSobj.Queries
				job.Tables = DFSobj.Tables
				job.IsDist = true
			default:
				log.Println("Error with config file: ",config,". Please provide a valid configuration type: RDB, DFS")
				continue
		}

		runnow := []qt.Query{}
		for _,query := range job.Queries {
			cron := query.Cron
			if qt.CheckTime(cron) {
				runnow = append(runnow,query)
			}
		}

		if len(runnow) > 0 {
			job.Queries = runnow
			jobs = append(jobs,job)
		}
	}

	return jobs
}

// sends the jobs via rpc connection to the receiver process
func sendToReceiver(jobs []qt.Job) {
	client, err := rpc.DialHTTP("tcp", "localhost:8675")
	if err != nil {
		log.Fatal(err)
	}

	for _,job := range jobs {
		var reply bool
		err = client.Call("Receiver.SendJob", &job, &reply)
		if err != nil {
			log.Fatal(err)
		}
	}
}


