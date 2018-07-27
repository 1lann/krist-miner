# krist-miner
The second fastest open source CPU Krist miner, written in Go.

## Performance
This miner is up to 3.5x faster than [YTCI Krist Miner](https://github.com/Yevano/ytci-krist-miner/).
It is at least 2.5x faster than YTCI Krist Miner.

On my Macbook Pro 15-inch Retina (Mid 2014) which has an Intel Core i7-4870HQ, running with 8 processes gives me speeds around 18 MH/s.

This miner has a very light memory footprint of < 3 MB. Whereas YTCI miner requires a Java VM which has a much larger overhead.

## Binaries
Binaries can be found under [releases](https://github.com/1lann/krist-miner/releases).

## Usage
Execute the binary in a terminal or command line for help.

## Donations
I don't even know why I'm bothering to put this section here, but if for some reason you would like to send me some virtual currency that has almost no value, feel free to send some KST to me for whatever reason.
```
k3be4p30lb
```

## GPU Miner?
A GPU miner written for OpenCL is in the works, it may or may not ever be completed.

## License
This krist-miner is licensed under the MIT license. This not apply to the code taken from [Minio's SHA-256 SIMD implementation](https://github.com/minio/sha256-simd), which is licensed under the Apache 2 license,
and such files subject to the Apache 2 license are noted in their headers.

No modifications were made to files from Minio's SHA-256 SIMD implementation.

```
The MIT License (MIT)

Copyright (c) 2016 Jason Chu (1lann)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

Files from Minio's SHA-256 implementation are noted at the top, and are licensed under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0). Here is the list of modifications made:
* Removed unnecessary code for Krist mining.
* All `.s` files are left untouched.
* All files in the cpu-miner/cpuid folder are left untouched.
