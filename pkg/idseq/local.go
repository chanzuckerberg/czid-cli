package idseq

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/metadata"
)

var inputExp = regexp.MustCompile(`\.(fasta|fa|fastq|fq)(\.gz)?$`)

func IsInput(path string) bool {
	return inputExp.MatchString(path)
}

var sampleNameExp = regexp.MustCompile(`(_R[12]|_R[12]_001)?\.(fasta|fa|fastq|fq)(\.gz)?$`)

func ToSampleName(path string) string {
	return sampleNameExp.ReplaceAllString(filepath.Base(path), "")
}

var r1Exp = regexp.MustCompile(`_R1(_001)?\.(fasta|fa|fastq|fq)(\.gz)?$`)

func IsR1(path string) bool {
	return r1Exp.MatchString(path)
}

var r2Exp = regexp.MustCompile(`_R2(_001)?\.(fasta|fa|fastq|fq)(\.gz)?$`)

func IsR2(path string) bool {
	return r2Exp.MatchString(path)
}

type SampleFiles struct {
	R1     string
	R2     string
	Single string
}

func SamplesFromDir(directory string, verbose bool) (map[string]SampleFiles, error) {
	pairs := make(map[string]SampleFiles)
	if dir, err := os.Stat(directory); err != nil {
		return pairs, err
	} else if !dir.IsDir() {
		return pairs, fmt.Errorf("path %s must be a directory", directory)
	}

	err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if match := IsInput(path); match {
			sampleName := ToSampleName(path)
			sampleFiles := pairs[sampleName]

			if IsR1(path) {
				if sampleFiles.Single != "" {
					return fmt.Errorf("found R1 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.Single)
				}
				if sampleFiles.R1 != "" {
					return fmt.Errorf("found multiple R1 files for sample '%s': %s, %s", sampleName, path, sampleFiles.R1)
				}

				if verbose {
					fmt.Printf("detected R1 sample file for sample: %s at path %s\n", sampleName, path)
				}

				sampleFiles.R1 = path
			} else if IsR2(path) {
				if sampleFiles.Single != "" {
					return fmt.Errorf("found R2 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.Single)
				}
				if sampleFiles.R2 != "" {
					return fmt.Errorf("found multiple R2 files for sample '%s': %s, %s", sampleName, path, sampleFiles.R2)
				}

				if verbose {
					fmt.Printf("detected R2 sample file for sample: %s at path %s\n", sampleName, path)
				}

				sampleFiles.R2 = path
			} else {
				if sampleFiles.R1 != "" {
					return fmt.Errorf("found R1 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.R1)
				}
				if sampleFiles.R2 != "" {
					return fmt.Errorf("found R2 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.R2)
				}
				if sampleFiles.Single != "" {
					return fmt.Errorf("found multiple single end files for sample '%s': %s, %s", sampleName, path, sampleFiles.Single)
				}

				if verbose {
					fmt.Printf("detected single sample file for sample: %s at path %s\n", sampleName, path)
				}

				sampleFiles.Single = path
			}
			pairs[sampleName] = sampleFiles
		}
		return err
	})
	for sampleName, pair := range pairs {
		if verbose {
			fmt.Printf("detected sample: %s\n", sampleName)
		}
		if pair.R1 != "" && pair.R2 == "" {
			return pairs, fmt.Errorf("found R1 but not R2 for sample '%s': %s", sampleName, pair.R1)
		}
		if pair.R1 == "" && pair.R2 != "" {
			return pairs, fmt.Errorf("found R2 but not R1 for sample '%s': %s", sampleName, pair.R2)
		}
	}
	return pairs, err
}

func GeoSearchSuggestions(samplesMetadata *metadata.SamplesMetadata) error {
	for sampleName, metadata := range *samplesMetadata {
		suggestion, err := GetGeoSearchSuggestion(metadata.collectionLocation, metadata.IsHuman())
		if err != nil {
			return err
		}
		if suggestion != (GeoSearchSuggestion{}) {
			metadata := (*samplesMetadata)[sampleName]
			metadata.CollectionLocation = &suggestion
			(*samplesMetadata)[sampleName] = metadata
		}
	}
	return nil
}
