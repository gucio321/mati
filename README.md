## Intro
A simple-purpose program dedicated for parsing
data from [halmar.pl](https://halmar.pl) website.

## Usage

download file from the latest Release, or (if you have Go installed) just run `go install github.com/gucio321/mati@latest`.

then run the following command:
```sh
mati -url <your URL> -dir <Output directory for your photos>
```

The programm will output all necessary data to the console and write images in `-dir`.
If you don't specify `-dir` or `-url`, the programm will prompt you for them.
