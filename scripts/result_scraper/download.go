package main

import (
	"errors"
	"fmt"
	"github.com/Eitol/verificador_elecciones2024_ve/pkg/indexcache"
	"github.com/Eitol/verificador_elecciones2024_ve/pkg/results"
	"log"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func scrape() {
	recaptcha := "03AFcWeA4OpGfWBmjCBXqwba_cDsjnRaPaLn76ridRM6ZEDykEZ5Hxb18X0pMriTPrqWC9VLkspD0IXARqujgw3gR1sX7QnCk3vrPtMdD3fM_j1y_CMdzdingL1JJmWcgeH_0U8vtKUgS866DQ4HChfh0KqTXcKpAMxPjfJgg46VpFJtLQ7H900D8wA3YopmewYVYGXgmtHD0ZxhLmIXAXXfNfVOuTmda_ebw-rVWeTO8r0KWUzOimdYki9UPyumNR_sAaw7ndf-94mBwMLqLSMXav-yjLHobKHQE6Rqu4a9u3MfDh0c1JroTKMqmkXdsji1HQXghxc4w8-9sGlm9hqWIce-Zu-C4d8jwud2mHJcariT5yYmiiPSEakd_rk0RyC3AvIfCd3IQutYiwlCdKVCxQaoiTUixwSY2y1PSreSIOW86PH3qpcg0iVg1tz8yQyVE459LmuTdOWMnXGwj6zuqVrKYsAehL807j-00dUwX66tMwv0Gxjp-XOut4IUABp2SXU0v67oC0pFR_ru_kAo-hQU6VL2wsj0gxvixe_Z771LR25doXOmjGgWG6Anl0M0yiN56goaBhRyao31ktlpJWUHcFS3tS054twTq3VNQAYOXxkkEmy62kf9zFtyvnSsD_gdmw8rD6h9UJ_9cF4k3w4y9qlBJDo9QAVy8VBgCnx-xZx6n-JPDftbEy64wc528wHDQZRWnQMzPnQ_54QiosiVmbXrrs1kR0TuM2GRU5leRpl7vJ4ePw7rlwKv62LkBdlXYtTn1owqUJAvz57sAJWb1goNp7SexkG_fO_6oO5wOgHRue3ZTz1NVfVkBx9VMndkUmc5MO"
	initialDocID := 15_000_000
	finalDocID := 25_000_000
	dirPath := filepath.Join("assets", "results")
	cacheFilePath := filepath.Join("assets", "results_index_cache.json")
	fmt.Printf("reading cache...")
	resultStore, err := results.NewJsonFileResultStore(dirPath)
	fmt.Printf("read cache done\n")
	if err != nil {
		log.Panicf("error creating result store: %v", err)
	}
	resultRepo := results.NewResultsRepo(resultStore, recaptcha)
	indexCache := indexcache.NewIndexCache(cacheFilePath)
	cachedIndex, err := indexCache.GetLatest()
	if err == nil {
		initialDocID, err = strconv.Atoi(cachedIndex.Latest)
		if err != nil {
			log.Panicf("error converting cached index to int: %v", err)
		}
	}

	nworkers := 100
	jobChan := make(chan int, nworkers)
	wg := &sync.WaitGroup{}
	count := 1
	go func() {
		for docID := initialDocID; docID <= finalDocID; docID++ {
			jobChan <- docID
			count++
			if count%1000 == 0 {
				fmt.Printf("sent docID %d\n", docID)
			}
		}
	}()

	for i := 0; i < nworkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for docID := range jobChan {
				err = indexCache.SetLatest(strconv.Itoa(docID))
				if err != nil {
					log.Panicf("error setting latest index: %v", err)
				}
				result, err := resultRepo.FindByDocID(docID)
				if result == nil || errors.Is(err, results.ErrResultNotFound) {
					//fmt.Println("result not found for docID", docID)
					continue
				}
				if result.IsUpdated {
					fmt.Printf("updated result for docID %d\n", docID)
				}
				if !result.IsCached {
					fmt.Printf("\n\ndownloaded result for docID %d\n\n\n", docID)
				}
			}
		}()
	}
	wg.Wait()
	fmt.Printf("\n\nfinished scraping results from docID %d to %d\n", initialDocID, finalDocID)
}

func main() {
	for {
		// recover from panic
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recovered from panic")
			}
		}()
		scrape()
		// sleep for a while before next run (optional)
		time.Sleep(1 * time.Minute)
	}
}
