# OPTIMISATIONS LIST OUT OF DATE

This is the optimisations present in krist-miner v0.4 and below. krist-miner v1.0 has even better optimisations, and I'm too lazy to update this document right now.

# Optimisations in krist-miner

This document will detail the optimisations which allows my miner to be faster than Yevano's Krist Miner.

## The Permuter (or Nonce Generator)

Like all Krist miners, we need to be able to generate a nonce to hash. In this document I will call this the "permuter".

Here's Yevano's permuter simplified:

```java
long nonce = 0;
while (true) {
	String nonceStr = Long.toString(nonce, 36);
	nonce++;
	...
}
```

Do you see the slow down in the code? Let's take a look at `Long.toString` shall we?

```java
public static String toString(long i, int radix) {
    if (radix < Character.MIN_RADIX || radix > Character.MAX_RADIX)
        radix = 10;
    if (radix == 10)
        return toString(i);
    char[] buf = new char[65];
    int charPos = 64;
    boolean negative = (i < 0);

    if (!negative) {
        i = -i;
    }

    while (i <= -radix) {
        buf[charPos--] = Integer.digits[(int)(-(i % radix))];
        i = i / radix;
    }
    buf[charPos] = Integer.digits[(int)(-i)];

    if (negative) {
        buf[--charPos] = '-';
    }

    return new String(buf, charPos, (65 - charPos));
}
```

That's quite a lot of instructions that needs to be executed on every iteration don't you think? You can also think of it as being wasteful in the fact that the ENTIRE number needs to be converted to the string, when in between each iteration, usually only the last character changes.

Now let's compare it to what I have in my miner:

```go
for {
	nonce = permalgo.Next()
	...
}
```

Here's `permalgo.Next()`

```go
type generator struct {
	lastString []byte
}

func (g *generator) Next() []byte {
	incrementString(g.lastString)
	return g.lastString
}

func incrementString(text []byte) {
	for place := len(text) - 1; place >= 0; place-- {
		if text[place] < 'z' {
			text[place] = text[place] + 1
			// Note that the function mainly terminates here
			return
		} else {
			text[place] = 'a'
		}
	}
}
```

That's it, a nice and fast ASCII permutator, little instructions wasted as it returns as soon as it can. In most cases, that's the first character. Instead of re-creating the entire string like what happens in Yevano's.

#### Reusing Memory

But wait! There's another optimisation in there already, let's make a more naive `incrementString` function:

```go
func incrementString(text []byte) []byte {
	for place := len(text) - 1; place >= 0; place-- {
		if text[place] < 'z' {
			text[place] = text[place] + 1
			return text
		} else {
			text[place] = 'a'
		}
	}
	// Must return here
	return text
}
```

It might not look that different, but there is a signficiant difference in how this code runs, compared to the one above. In this one, I'm returning an array of bytes (`[]byte`) whereas in the first one, I'm writing new data into the same array, as arrays are pass by reference.

What does your computer do differently with this new one? It needs to allocate memory for the new array that you're returning. You might be thinking "oh but that's only a couple of more bytes of allocated memory". Not quite, the garbage collector is actually responsible for cleaning up memory in this case. This also means that it is not necessarily cleaned up instantly at the end of the scope, meaning memory use can actually build up. By permuting the nonce in an existing array you are not only reducing memory usage, but also reducing pressure on the garbage collector.

## The SHA Algorithm?

Believe it or not, there's nothing special to my krist-miner's SHA256 algorithm. In fact I'm using the built in standard library one. But there are hidden optimisations that you can do "surrounding" the SHA algorithm call.

### Invoking the SHA Algorithm

Let's take a look at how the SHA256 algorithm is called in Yevano's miner:

```java
SHA256.digest((addrWBlock + nonceStr).getBytes(StandardCharsets.UTF_8))
```

Looks pretty typical, right? Now let's look at my miner's:

```go
header = append(header[:headerLen], nonce...)
if sha2algo.Sum256NumberCmp(header, maxWork) {
	...
}
```


#### Look Ma, No Conversions!

The first thing to notice is that I don't need to convert a string to bytes! I actually store the permutable nonce and header (what's called `addrWBlock` in Yevano's miner) as a raw bytes already. That saves a few instructions to not have to convert between data structures, but remember that every instruction counts!

#### Reusing Memory Part 2

Another optimisation I do is that the nonce is added to the header in-place. Compare these two operations from Yevano's miner and mine:

```java
addrWBlock + nonceStr
```

```go
header = append(header[:headerLen], nonce...)
```

For Yevano's miner, Java needs to allocate new memory to store `addrWBlock + nonceStr` and is stored seperately from addrWBlock and nonceStr. This needs to be done for every iteration of the loop.

My method is different, instead of requiring the language from allocating new memory, I use existing memory by appending the nonce to the end of the header, and then passing the same piece of memory into the SHA256 algorithm. This allows for efficient memory use and reduces load on the memory allocator. As you can probably tell too, the header is trimmed (`header[:headerLen]`) each time a nonce is added on. Go's escape analysis doesn't actually work for objects with append() on them. So by storing the data back into `header` instead of a new variable, I also significantly reduce pressure on the garbage collector.

### Calculating the Hash "Value" or "Number"

As part of the Krist specification, you need to take the last 6 bytes of the hash, and turn it into a number which is to be compared to the `work`. If it is less than the work, then it can be submitted as the solution and you are awarded KST.

Let's take a look at how Yevano's miner does it:

```java
SHA256.hashToLong(SHA256.digest(...))
```

Here's the code for hashToLong:

```java
public static long hashToLong(byte[] hash) {
	long ret = 0;
	for (int i = 5; i >= 0; i--) {
		ret += (hash[i] & 0xFF) * Math.pow(256, 5 - i);
	}
	return ret;
}
```

I'm not sure what the `& 0xFF` is for, but the only immediate optimisation you can do with this is to use a bit shift instead of `Math.pow(256, 5 - i)`.

Let's take a look at my miner:

```go
if sha2algo.Sum256NumberCmp(header, maxWork) {
	...
}
```

So my SHA256 function takes in the work value, that's different. Here's the code for Sum256NumberCmp:

```go
func (g *generator) Sum256NumberCmp(data []byte, work int64) bool {
    result := sha256.Sum256(data)

    value := int64(result[5])
    if value > work {
        return false
    }

    value += int64(result[4]) << (8 * 1)
    if value > work {
        return false
    }

    value += int64(result[3]) << (8 * 2)
    if value > work {
        return false
    }

    value += int64(result[2]) << (8 * 3)
    if value > work {
        return false
    }

    value += int64(result[1]) << (8 * 4)
    if value > work {
        return false
    }

    value += int64(result[0]) << (8 * 5)
    if value > work {
        return false
    }

    return true
}
```

Instead of calculating the entire value of the hash, I hardcode the conversions and short circuit it. Why bother converting every byte when the first or second byte can tell 95% of the time you whether the hash value will be less than the work? This allows for some reduction in instructions.

## Garbage Collection

As a result of the memory reuse and careful design of my miner, I'm actually capable of almost completely turning off the Go garbage collector. Everything can be stack allocated and everything else that can't be, is reused. The garbage collector actually still needs to run to clean up a few kilobytes of memory caused by the HTTP library, but I can manually invoke it, and I do once every minute. Which means the garabge collector takes up less than 0.001% of the execution time. That adds a bit of a performance boost and allows the miner to run even faster than Yevano's miner, as his still needs the garbage collector to be enabled.

## Conclusion

It's through all these optimisations, whether big or small, which allows my miner to run 1.3-1.8x faster than Yevano's. It's amazing how such small details that you often look over, such as the use of `Float.toString` can actually have a significant impact on performance. Because when it comes to something as intensive as mining a cyrptocurrency, that you need to run as fast as you possibly can, every instruction saved counts.

Remember, the slowest part of the mining process is actually calculating the SHA256 hash, and I'm not even touching that, yet my miner can run 1.3-1.8x faster. So purely through reducing overhead, my miner is able to boost performance quite significantly.

I had a great time writing my miner, and thinking about all the optimisations I could make. In fact, I was trying many differnet methods as you can tell by looking at the code, the SHA256 algorithm and "permuter" are actually modular, and I tried many different algorithms to find the fastest combination. I was quite surprised to find that the fastest SHA256 algorithm was actually the one in the Go standard library!

Thank you for reading this. I hope this was worth reading and you learnt something new from it. Happy coding!
