# IDSeq CLI

This project is under development and not yet stable.

A **C**ommand **L**ine **I**nterface for [IDSeq](https://idseq.net/). 

Rewrite of https://github.com/chanzuckerberg/idseq-cli, work in progress.

## Getting Started

### Installation

Currently only binaries are available. More types of packages coming soon.

#### Linux

There are lots of options to install on Linux.

##### Debian Distributions (ex. Debian, Ubuntu, Linux Mint)

1. Download our [latest .deb package](https://github.com/chanzuckerberg/idseq-cli-v2/releases/latest/download/idseq-cli-v2_linux_amd64.deb).
1. (Alternatively) Download with curl: `curl -L https://github.com/chanzuckerberg/idseq-cli-v2/releases/latest/download/idseq-cli-v2_linux_amd64.deb -o idseq-cli-v2_linux_amd64.deb`
1. Install the package: `sudo dpkg -i idseq-cli-v2_linux_amd64.deb`
1. (Optional) Remove the package file `rm idseq-cli-v2_linux_amd64.deb`


##### Fedora Distributions (ex. Centos, RHEL)

1. Download our [latest .rpm package](https://github.com/chanzuckerberg/idseq-cli-v2/releases/latest/download/idseq-cli-v2_linux_amd64.rpm)
1. (Alternatively) Download with curl: `curl -L https://github.com/chanzuckerberg/idseq-cli-v2/releases/latest/download/idseq-cli-v2_linux_amd64.rpm -o idseq-cli-v2_linux_amd64.rpm`
1. Install the package: `sudo rpm -i s3parcp.rpm`
1. (Optional) Remove the package file: `rm idseq-cli-v2_linux_amd64.rpm`

##### Other Linux: Install via Homebrew for Linux

1. Make sure you have [Homebrew for Linux](https://docs.brew.sh/Homebrew-on-Linux)
1. Add the chanzuckerberg tap: `brew tap chanzuckerberg/tap`
1. Install the package: `brew install s3parcp`

##### Other Linux: Without Homebrew

Follow the instructions for [installing from binaries](#from-binaries).

#### MacOs

idseq-cli-v2 is available via homebrew and natively supports Apple Silicon! (via the `darwin_arm64` binary)

##### Install via Homebrew

1. Make sure you have [Homebrew for Linux](https://docs.brew.sh/Homebrew-on-Linux)
1. Add the chanzuckerberg tap: `brew tap chanzuckerberg/tap`
1. Install the package: `brew install s3parcp`

##### Without Homebrew

Follow the instructions for [installing from binaries](#from-binaries).

#### Windows

Follow the instructions for [installing from binaries](#from-binaries).

#### From Binaries

1. Navigate to our [latest release](https://github.com/chanzuckerberg/idseq-cli-v2/releases/latest).
1. Available binaries will be under the `Assets` tab as archives (`.tar.gz` for Linux + MacOS, `.zip` for Windows)
1. Download appropriate archive for your operating system and architecture
1. Unzip the file, the `idseq` executable will be inside

    Linux + MacOS: `tar -xf path/to/archive.tar.gz`

    Windows: `expand-archive -path 'c:\path\to\archive.zip' -destinationpath '.\idseq-cli-v2'`

1. Files with shell completions are also inside so you can move them to the appropriate place for your shell


Note on MacOS: Currently we don't sign our binary so you will need to manually remove the quarentine attribute from the binary: `sudo xattr -d com.apple.quarantine path/to/binary`


### Usage

First log in with your IDSeq account:

```bash
idseq login
```

You will be prompted to log in with your IDSeq account via the web.

Accept the user agreement:

```bash
idseq accept-user-agreement
```

This will print the user agreement and prompt you for your agreement.

Optionally, you can create a metadata CSV file for your sample. You can skip this step and specify your metadata with command line flags. For instructions on creating this file see:

- Instructions: https://idseq.net/metadata/instructions
- Metadata dictionary and supported host genomes: https://idseq.net/metadata/dictionary
- Metadata CSV template: https://idseq.net/metadata/metadata_template_csv

Upload a single sample to the short read MNGS pipeline:

```bash
idseq short-read-mngs upload-sample \
  -p 'Project Name' \
  -s 'Sample Name' \
  --metadata-csv your_metadata.csv \
  -m 'Metadata Name=Metadata Value' \
  your_sample_R1.fastq.gz your_sample_R2.fastq.gz
```

Note, R2 is optional if uploading single end reads. Supported file types: (.fastq/.fq/.fasta/.fa/.fastq.gz/.fq.gz/.fasta.gz/.fa.gz). If you would like to specify your metadata entirely with `-m` flags you don't need to include a `--metadata-csv`. If you have specified all of your metadata in the metadata csv you don't need to include any `-m` flags. `-m` flags override metadata from the csv.

## Differences from version 1

- Resume uploads that have been interrupted
- Faster uploads (on systems with high bandwidth) due to multithreading
- Distributed as a single binary for easier installation that doesn't rely on dependencies on the user's machine
- Shell completion support
- Structured with commands and subcommands to make room for future functionality
- Log in in the web via your IDSeq account instead of using a static token
- Critical bugfixes
- Uploads without the need for user prompts
- Supports configuration files and environment variable configuration instead of relying solely on flags

## Contributing

This project is not seeking contributions at this time. It is tighly coupled to the IDSeq Web App, it's features, it's APIs, and it's development goals. Please feel free to raise issues for feature requests or bugs.

This project adheres to the Contributor Covenant [code of conduct](https://www.contributor-covenant.org/). By participating, you are expected to uphold this code. Please report unacceptable behavior to opensource@chanzuckerberg.com.

## Reporting Security Issues

Please note: If you believe you have found a security issue, please responsibly disclose by contacting us at security@idseq.net.

See [SECURITY.md](SECURITY.md) for more information.
