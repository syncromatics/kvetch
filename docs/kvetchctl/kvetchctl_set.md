## kvetchctl set

Set values by key

### Synopsis

Sets values for keys

If both a key and value are specified as arguments, the key will be set with the given value.
If only a key is specified as an argument, the key will be set with the value read from STDIN.
If neither a key nor value is specified as an argument, the keys and values will be read from JSON objects from STDIN.

```
kvetchctl set [flags] [key] [value]
-or- set [flags] [key]
-or- set [flags]
```

### Options

```
  -e, --endpoint string     Kvetch instance to connect to (required)
  -h, --help                help for set
  -o, --output string       Set the output format (simple, json) (default "simple")
      --ttl duration        Set the time-to-live for each key (optional)
  -t, --value-type string   Set the type of value in the output (string, bytes, json) (default "string")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [kvetchctl](kvetchctl.md)	 - Command line interface for interacting with Kvetch

###### Auto generated by spf13/cobra on 15-Jun-2020
