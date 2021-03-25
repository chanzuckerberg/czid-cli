# IDSeq CLI

This project is under development and not yet stable.

A **C**ommand **L**ine **I**nterface for [IDSeq](https://idseq.net/). 

Rewrite of https://github.com/chanzuckerberg/idseq-cli, work in progress.

## Getting Started

Currently this project is not ready for use. Getting Started instructions coming soon.

### Installation

Currently only binaries are available. More types of packages coming soon.

#### From Binaries

1. Navigate to our [latest release](https://github.com/chanzuckerberg/idseq-cli-v2/releases/latest).
1. The binaries will be under the `Assets` tab
1. Download the one for your operating system and architecture
1. Unzip the file, the `idseq` executable will be inside

### Usage

First log in with your IDSeq account:

```
idseq login
```

Accept the user agreement:

```
idseq accept-user-agreement
```

Optionally, you can create a metadata CSV file for your sample. You can skip this step and specify your metadata with command line flags. For instructions on creating this file see:

  - Instructions: https://idseq.net/metadata/instructions
  - Metadata dictionary and supported host genomes: https://idseq.net/metadata/dictionary
  - Metadata CSV template: https://idseq.net/metadata/metadata_template_csv

Upload a single sample to the short read MNGS pipeline:

```
idseq short-read-mngs upload-sample -s 'Project Name' -s 'Sample Name' --metadata-csv your_metadata.csv -m 'Metadata Name=Metadata Value' your_sample_R1.fastq.gz your_sample_R2.fastq.gz
```

Note, R2 is optional if uploading single end reads. Supported file types: (.fastq/.fq/.fasta/.fa/.fastq.gz/.fq.gz/.fasta.gz/.fa.gz). If you would like to specify your metadata entirely with `-m` flags you don't need to include a `--metadata-csv`. If you have specified all of your metadata in the metadata csv you don't need to include any `-m` flags. `-m` flags override metadata from the csv.

## Contributing

This project is not currently accepting contributions at this time.

This project adheres to the Contributor Covenant [code of conduct](https://www.contributor-covenant.org/). By participating, you are expected to uphold this code. Please report unacceptable behavior to opensource@chanzuckerberg.com.

## Reporting Security Issues

Please note: If you believe you have found a security issue, please responsibly disclose by contacting us at security@idseq.net.

See [SECURITY.md](SECURITY.md) for more information.
