package idseq

type Metadata = map[string]string
type SamplesMetadata = map[string]Metadata

func ToValidateForm(m SamplesMetadata) validationMetadata {
	headerIndexes := map[string]int{"Sample Name": 0}
	vM := validationMetadata{
		Headers: []string{"Sample Name"},
		Rows:    make([][]string, len(m)),
	}

	for sampleName, row := range m {
		validatorRow := make([]string, len(vM.Headers))
		validatorRow[0] = sampleName
		for name, value := range row {
			headerIndex, seenHeader := headerIndexes[name]
			if !seenHeader {
				vM.Headers = append(vM.Headers, name)
				headerIndexes[name] = len(headerIndexes)
				validatorRow = append(validatorRow, value)
			} else {
				validatorRow[headerIndex] = value
			}
		}
		vM.Rows = append(vM.Rows, validatorRow)
	}
	return vM
}
