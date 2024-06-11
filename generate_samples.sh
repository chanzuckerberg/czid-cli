#!/bin/bash

# Paths to the original fastq files
SOURCE="/Users/ovalenzuela/Documents/cli_files_501/omar_test_ds.1.fq"

# Directory to store the generated samples
OUTPUT_DIR="./samples"

# Create the output directory if it does not exist
mkdir -p "${OUTPUT_DIR}"

# Generate 501 unique samples
for i in $(seq 1 501); do
    # Copy and rename the R1 and R2 files to the sample directory
    cp "${SOURCE}" "${OUTPUT_DIR}/omar_test_ds.${i}.fq"
done

echo "501 unique samples have been generated in the '${OUTPUT_DIR}' directory."

