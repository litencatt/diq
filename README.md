# diq
- Simple tool to search DNS records.
- Support output format
  - STDOUT
  - JSON

## Usage
[Dowonload latest release](https://github.com/litencatt/diq/releases/latest)
```
$ diq google.com
@8.8.8.8
A       2404:6800:400a:808::200e
A       216.58.196.238
```

Default settings saved in `$HOME/.diq.yml`
```
$ cat ~/.diq.yml 
nameservers:
  - 8.8.8.8
qtypes:
  - A
```

### flags
`--config string   config file (default is $HOME/.diq.yml)`
```
$ cat /path/to/.diq.yml 
nameservers:
  - 8.8.8.8
  - 1.1.1.1
qtypes:
  - A
  - NS
  
$ diq --config /path/to/.diq.yml google.com
google.com
@8.8.8.8
A       2404:6800:400a:809::200e
A       172.217.161.206
NS      ns2.google.com.
NS      ns4.google.com.
NS      ns1.google.com.
NS      ns3.google.com.

@1.1.1.1
A       2404:6800:400a:80c::200e
A       172.217.161.238
NS      ns3.google.com.
NS      ns4.google.com.
NS      ns2.google.com.
NS      ns1.google.com.
```

`-q, --qtype string    lookup query types. (e.g. -q a,mx)`
```
$ diq google.com -q a,mx,ns
google.com
@8.8.8.8
A       2404:6800:400a:808::200e
A       216.58.197.14
MX      aspmx.l.google.com.
MX      alt1.aspmx.l.google.com.
MX      alt2.aspmx.l.google.com.
MX      alt3.aspmx.l.google.com.
MX      alt4.aspmx.l.google.com.
NS      ns3.google.com.
NS      ns1.google.com.
NS      ns2.google.com.
NS      ns4.google.com.
```
### output JSON
```
$ diq google.com -q a,mx,ns --format json | jq .
{
  "Domains": [
    {
      "DomainName": "google.com",
      "Result": [
        {
          "Nameserver": "@8.8.8.8",
          "Records": [
            {
              "Type": "A",
              "Record": [
                "2404:6800:400a:807::200e",
                "216.58.196.238"
              ]
            },
            {
              "Type": "MX",
              "Record": [
                "aspmx.l.google.com.",
                "alt1.aspmx.l.google.com.",
                "alt2.aspmx.l.google.com.",
                "alt3.aspmx.l.google.com.",
                "alt4.aspmx.l.google.com."
              ]
            },
            {
              "Type": "NS",
              "Record": [
                "ns2.google.com.",
                "ns4.google.com.",
                "ns1.google.com.",
                "ns3.google.com."
              ]
            }
          ]
        }
      ]
    }
  ]
}
```
