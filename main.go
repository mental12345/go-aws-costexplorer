package main

import (
  "fmt"
  "os"
  "encoding/csv"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/service/costexplorer"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/urfave/cli"
  "time"
  "strconv"
)

func timeNow(x int) (string, string) {
  dt := time.Now()
  now := dt.Format("2006-01-02")
  month := dt.AddDate(0, 0, -x).Format("2006-01-02")
  return now, month
}

func costAWS(time int) ([][]string){
  sess, _ := session.NewSession()
  svc := costexplorer.New(sess)
  now, monthBefore := timeNow(time)

  input := &costexplorer.GetCostAndUsageInput{
    Granularity: aws.String("DAILY"),
    TimePeriod: &costexplorer.DateInterval{
      Start: aws.String(monthBefore),
      End: aws.String( now ),
    },
    Metrics: []*string{
      aws.String("UNBLENDED_COST"),
    },
    GroupBy: []*costexplorer.GroupDefinition{
      {
        Type: aws.String("DIMENSION"),
        Key: aws.String("SERVICE"),
      },
    },
  }
  
  req, resp := svc.GetCostAndUsageRequest(input)
  
  err := req.Send()
  if err != nil {
    fmt.Println(err)
  }
  var resultsCosts [][]string
 
  for _, results := range resp.ResultsByTime {
    startDate := *results.TimePeriod.Start
    for _, groups := range results.Groups{
      for _, metrics := range groups.Metrics{
        info := []string{startDate, *groups.Keys[0], *metrics.Amount}
        resultsCosts = append(resultsCosts, info)
        
      }
    }
  } 
  return resultsCosts
}

func writeFile( fileName string, results [][]string ) {
  file, _ := os.Create(fileName)
  csvwriter := csv.NewWriter(file)
  for _, row := range results {
    _ = csvwriter.Write(row)
  }
  
  csvwriter.Flush()
  file.Close()
  
}

func main() {  
  app := cli.NewApp()
  app.Name = "A simple cost analyzer for aws"
  app.Author = "Ramon Esparza @ DigitalOnUs"
  app.Version = "0.1"
  app.Commands = []cli.Command{
    {
      Name: "gen_cost",
      Aliases: []string{"gc"},
      Usage: "It gives the cost on aws account in a period of time. Arguments: * number of days to calculate the aws costs, * name of the oputput file",
      Action: func(c *cli.Context) {
        timeinDays := c.Args().Get(0)
        fileName := c.Args().Get(1)
        days, _ := strconv.Atoi(timeinDays)
        writeFile(fileName, costAWS(days))
      },
    },
  }
  err := app.Run(os.Args)
  if err != nil {
    fmt.Println("Chaos error")
  }
}
