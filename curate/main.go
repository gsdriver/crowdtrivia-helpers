package main

import (
  "fmt"
  "context"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type MyEvent struct {
  Records interface{} `json:"Records"`
}

type MyItem struct {
  id string
  token string
  question string
  answer string
  incorrect []string
  upVote int
  downVote int
  created int64
}

const QuestionsPerFile = 3

func HandleRequest(ctx context.Context, event MyEvent) (string, error) {
  // We'll create three files of questions
  // One will be all new questions
  // The second will be a mix of new and recently curated
  // The third will be random questions from the DB

  // Start by reading in all entries from the DB
  sess, err := session.NewSession(&aws.Config{
    Region: aws.String("us-east-1")},
  )

  // Create DynamoDB client
  svc := dynamodb.New(sess)
  input := &dynamodb.ScanInput{
    TableName: aws.String("TriviaQuestions"),
  }

  // For now do a single scan; probably need to beef this up
  // as the number of questions stored grows
  result, err := svc.Scan(input)
  if err != nil {
    fmt.Println(err.Error())
    return "", err
  }

  // Sort the results into two sets as follows:
  // First set - not previously used (no up/down votes)
  // Second set - (30 - age in days) + up votes - down votes
  // Anything with an overall down vote of -5 or more gets dropped entirely
  var newQuestions []MyItem
  var otherQuestions []MyItem

  for _, i := range result.Items {
    item := MyItem{}
    err = dynamodbattribute.UnmarshalMap(i, &item)
    if err != nil {
      panic(fmt.Sprintf("Couldn't unmarshal record, %v", err))
    }

    if (item.upVote > 0) || (item.downVote > 0) {
      if !(item.downVote > (item.upVote + 5)) {
        // Add to other questions array
        otherQuestions = append(otherQuestions, item)
      }
    } else {
      // Put into newQuestions array
      newQuestions = append(newQuestions, item)
    }
  }

  return "OK", nil
}

func main() {
  lambda.Start(HandleRequest)
}
