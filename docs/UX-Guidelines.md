# gsctl UX guidelines

## Grammar

We use the `gsctl <verb> <subject>` grammar in all possible cases to make
commands discoverable and memorizable intuitively.

Examples:

    gsctl list clusters
    gsctl create keypair

## Command line flags

### Short vs. long form

Flags MUST have a long form. A long form is prepended with two dashes,
like `--verbose`.

Flags MAY have a short form in addition to the long form. The short form
is prefixed with only one dash and is only one letter long, like `-v`.

### Character set

Flags should only use the characters `a-z` and `-`, mainly to simplify input.

## Provider-specific options

Commands like `gsctl create cluster` will have provider-specific options, for
example, for setting the AWS EC2 instance types of worker nodes. These flags get
prefixxed with the provider name, e. g. `aws` and `azure`.

Examples:

    --aws-instance-type
    --aws-resource-tags
    --azure-machine-size
    --kvm-memory-size
    --kvm-cpu-cores
    --kvm-memory-size

## Colors

- The color *RED* is reserved for the error messages.

- The color *GREEN* is reserved for success messages.

- *CYAN* is reserved for:
  - table headers
  - verbatim strings, like URLs, in a text

- *YELLOW* is reserved for:
  - commands
  - feedback lines that are neither positive nor negative (e. g. "no clusters available" when listing clusters)
