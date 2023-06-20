package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

const (
	escaped          = "single''quote string"
	unescaped        = "single'quote string"
	storageAccountCS = "<real storage account>"
	azuriteCS        = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;"
)

func main() {
	sc, err := aztables.NewServiceClientFromConnectionString(azuriteCS, nil)
	if err != nil {
		panic(err)
	}
	client := sc.NewClient("testTable")
	client.CreateTable(context.Background(), nil)

	fmt.Println("testing escaped")
	test(client, escaped)

	// fmt.Println("testing unescaped")
	// test(client, unescaped)
}

func test(c *aztables.Client, s string) {
	if err := insertEntity(c, fmt.Sprintf("pk %s", s), fmt.Sprintf("rk %s", s), nil); err != nil {
		panic(err)
	}
	e, err := getEntity(c, fmt.Sprintf("pk %s", s), fmt.Sprintf("rk %s", s))
	if err != nil {
		panic(err)
	}
	for _, entity := range e {
		fmt.Println(entity)
	}
}

func insertEntity(c *aztables.Client, pk, rk string, props map[string]interface{}) error {
	myEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: pk,
			RowKey:       rk,
		},
		Properties: props,
	}
	marshalled, err := json.Marshal(myEntity)
	if err != nil {
		return err
	}
	_, err = c.UpsertEntity(context.Background(), marshalled, &aztables.UpsertEntityOptions{UpdateMode: aztables.UpdateModeMerge})
	if err != nil {
		tErr, ok := err.(*azcore.ResponseError)
		if !ok {
			return err
		}
		return tErr
	}
	return err
}

func getEntity(c *aztables.Client, pk, rk string) ([]*aztables.EDMEntity, error) {
	opts := &aztables.ListEntitiesOptions{
		Top:    to.Ptr(int32(1)),
		Filter: to.Ptr(fmt.Sprintf("PartitionKey eq '%s' and RowKey eq '%s'", pk, rk)),
	}
	pager := c.NewListEntitiesPager(opts)
	var entities []*aztables.EDMEntity
	for pager.More() {
		var (
			response aztables.ListEntitiesResponse
			err      error
		)
		response, err = pager.NextPage(context.Background())
		if err != nil {
			tErr, ok := err.(*azcore.ResponseError)
			if !ok {
				panic(err)
			}
			if !strings.Contains(tErr.Error(), "TableNotFound") {
				panic(err)
			}
			if _, err = c.CreateTable(context.Background(), nil); err != nil {
				panic(err)
			}
			response, err = pager.NextPage(context.Background())
			if err != nil {
				panic(err)
			}
		}
		for _, e := range response.Entities {
			var entity *aztables.EDMEntity
			err = json.Unmarshal(e, &entity)
			if err != nil {
				panic(err)
			}
			entities = append(entities, entity)
		}
	}
	return entities, nil
}
