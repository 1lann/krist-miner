# krist-miner
A fast CPU krist miner written in Go, designed to be modular, extendable and fast.

## Notice
This miner is very new, and unfortunately submitting solutions is untested due to the current state of Krist that
it's very difficult to mine any blocks. I'm not sure if the solutions this miner
finds are correct, and whether or not it can submit solutions without issues.

If you're able to test whether it submit solutions correctly, please let me know if it works or not.

## Performance
This miner is up to 1.8x faster than [YTCI Krist Miner](https://github.com/Yevano/ytci-krist-miner/).
It is at least 1.3x faster than YTCI Krist Miner.

On my Macbook Pro 15-inch Retina (Mid 2014) which has an Intel Core i7-4870HQ, running with 8 processes gives me speeds around 7 MH/s. The miner actually performs better or the same with 5 processes (not sure why), which also allows other applications to use more the CPU.

On my Raspberry Pi 2 Model B, running with 4 processes gives me speeds of around 300-320 kH/s.

This miner has a very light memory footprint of < 2.5 MB. Whereas YTCI miner requires a Java VM which has a
much larger overhead.

## Binaries
Binaries can be found under [releases](https://github.com/1lann/krist-miner/releases).

## Usage
Execute the binary in a terminal or command line for help.

## Donations
I don't even know why I'm bothering to put this section here, but if for some reason you would like to send me some virtual currency that has almost no value, feel free to send some KST to me for whatever reason: `k3be4p30lb`.

## GPU Miner?
A GPU miner written for OpenCL is in the works, it may or may not ever be completed.

## License
This krist-miner is licensed under the MIT license:

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
