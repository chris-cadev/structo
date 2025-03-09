package main

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/dsoprea/go-exif"
	log "github.com/dsoprea/go-logging"
)

func GetDateTaken(path string) (*time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		return nil, err
	}

	// Run the parse.
	im := exif.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()

	var dateTaken string

	visitor := func(fqIfdPath string, ifdIndex int, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (err error) {
		defer func() {
			if state := recover(); state != nil {
				err = log.Wrap(state.(error))
				log.Panic(err)
			}
		}()

		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)
		log.PanicIf(err)

		it, err := ti.Get(ifdPath, tagId)
		if err != nil {
			if log.Is(err, exif.ErrTagNotFound) {
				return nil
			}
			log.Panic(err)
		}

		// Check if the tag is DateTimeOriginal (Tag ID 0x9003)
		if it.Name == "DateTimeOriginal" {
			valueString, err := valueContext.FormatFirst()
			log.PanicIf(err)

			dateTaken = valueString
		}

		return nil
	}

	_, err = exif.Visit(exif.IfdStandard, im, ti, rawExif, visitor)
	if err != nil {
		return nil, err
	}
	layout := "2006:01:02 15:04:05"
	parsedTime, err := time.Parse(layout, dateTaken)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}
