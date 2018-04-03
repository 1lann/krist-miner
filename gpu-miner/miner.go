package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"unsafe"

	"github.com/1lann/sha256-simd"
	"github.com/go-gl/cl/v1.2/cl"
)

const (
	// Buffer size for communication to devices
	BufferSize = 4096
)

// KernelSource is the source code of the program we're going to run.
var KernelSource = `
// This file contains code for hashing and mining on OpenCL hardware

typedef uchar byte;

// union to convert 4 bytes to an int or vice versa
union byte_int_converter {
	uint val;
	byte bytes[4];
};

// macro so i can change it later
#define mult_add(a,b,c) (a * b + c)

// right rotate macro
#define RR(X, Y) rotate((uint)X, -((uint)Y))

// optimized padding macro
// takes a character array and integer
// character array is used as both input and output
// character array should be 64 items long regardless of content
// actual input present in character array should not exceed 55 items
// second argument should be the length of the input content
// example usage:
//	char data[64];
//	data[0] = 'h';
//	data[1] = 'e';
//	data[2] = 'l';
//	data[3] = 'l';
//	data[4] = 'o';
//	PAD(data, 5);
//	// data array now contains 'hello' padded
#define PAD(X, Y) X[63] = Y * 8; X[62] =  Y >> 5; X[Y] = 0x80;

// SHA256 macros
#define CH(x,y,z) bitselect(z,y,x)
#define MAJ(x,y,z) bitselect(x,y,z^x)
#define EP0(x) (RR(x,2) ^ RR(x,13) ^ RR(x,22))
#define EP1(x) (RR(x,6) ^ RR(x,11) ^ RR(x,25))
#define SIG0(x) (RR(x,7) ^ RR(x,18) ^ ((x) >> 3))
#define SIG1(x) (RR(x,17) ^ RR(x,19) ^ ((x) >> 10))

__constant uint K[64] = {
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
	0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
	0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
	0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2};

// SHA256 digest function - optimization pending
// takes a byte array of size 64 with 55 or fewer items, and writes
// hash to 32 item byte array.
// make sure to pass the input size as the second argument.
// example usage:
//	char data[64];
//  char hash[32];
//	data[0] = 'h';
//	data[1] = 'e';
//	data[2] = 'l';
//	data[3] = 'l';
//	data[4] = 'o';
//	digest(data, 5, hash);
//	// hash array now contains hash of 'hello'

static void digest(byte* data, uint inputLen, byte* hash) {
	/* init vars */
	union byte_int_converter h0, h1, h2, h3, h4, h5, h6, h7, temp;
	uint a, b, c, d, e, f, g, h, i, t1, t2, m[64] = {0};
	PAD(data, inputLen);
	/* init hash state */
	h0.val = 0x6a09e667;
	h1.val = 0xbb67ae85;
	h2.val = 0x3c6ef372;
	h3.val = 0xa54ff53a;
	h4.val = 0x510e527f;
	h5.val = 0x9b05688c;
	h6.val = 0x1f83d9ab;
	h7.val = 0x5be0cd19;
	/* transform */
#pragma unroll
	for (i = 0; i < 16; i++) {
		//m[i] = (data[mult_add(i,4,0)] << 24) | (data[mult_add(i,4,1)] << 16) | (data[mult_add(i,4,2)] << 8) | (data[mult_add(i,4,3)]);
		temp.bytes[3] = data[i*4];
		temp.bytes[2] = data[i*4+1];
		temp.bytes[1] = data[i*4+2];
		temp.bytes[0] = data[i*4+3];
		m[i] = temp.val;
	}
#pragma unroll
	for (i = 16; i < 64; ++i)
		m[i] = SIG1(m[i - 2]) + m[i - 7] + SIG0(m[i - 15]) + m[i - 16];
	a = h0.val;
	b = h1.val;
	c = h2.val;
	d = h3.val;
	e = h4.val;
	f = h5.val;
	g = h6.val;
	h = h7.val;
#pragma unroll
	for (i = 0; i < 64; ++i) {
		t1 = h + EP1(e) + CH(e,f,g) + K[i] + m[i];
		t2 = EP0(a) + MAJ(a,b,c);
		h = g;
		g = f;
		f = e;
		e = d + t1;
		d = c;
		c = b;
		b = a;
		a = t1 + t2;
	}
	h0.val += a;
	h1.val += b;
	h2.val += c;
	h3.val += d;
	h4.val += e;
	h5.val += f;
	h6.val += g;
	h7.val += h;
	/* finish */
	hash[0] = h0.bytes[3];
	hash[1] = h0.bytes[2];
	hash[2] = h0.bytes[1];
	hash[3] = h0.bytes[0];

	hash[4] = h1.bytes[3];
	hash[5] = h1.bytes[2];
	hash[6] = h1.bytes[1];
	hash[7] = h1.bytes[0];

	hash[8] = h2.bytes[3];
	hash[9] = h2.bytes[2];
	hash[10] = h2.bytes[1];
	hash[11] = h2.bytes[0];

	hash[12] = h3.bytes[3];
	hash[13] = h3.bytes[2];
	hash[14] = h3.bytes[1];
	hash[15] = h3.bytes[0];

	hash[16] = h4.bytes[3];
	hash[17] = h4.bytes[2];
	hash[18] = h4.bytes[1];
	hash[19] = h4.bytes[0];

	hash[20] = h5.bytes[3];
	hash[21] = h5.bytes[2];
	hash[22] = h5.bytes[1];
	hash[23] = h5.bytes[0];

	hash[24] = h6.bytes[3];
	hash[25] = h6.bytes[2];
	hash[26] = h6.bytes[1];
	hash[27] = h6.bytes[0];

	hash[28] = h7.bytes[3];
	hash[29] = h7.bytes[2];
	hash[30] = h7.bytes[1];
	hash[31] = h7.bytes[0];
}

static long hashToLong(byte* hash) {
	return hash[5] + (hash[4] << 8) + (hash[3] << 16) + ((long)hash[2] << 24) + ((long) hash[1] << 32) + ((long) hash[0] << 40);
}

__kernel void krist_miner_basic(
		__global const byte* address,	// 10 chars
		__global const byte* block,	// 12 chars
		__global const byte* prefix,	// 2 chars
		const long base,				// convert to 10 chars
		const long work,
		__global byte* output) {
	int id = get_global_id(0);
	long nonce = id + base;
	byte input[64];
	byte hashed[32];
#pragma unroll
	for (int i = 0; i < 10; i++) {
		input[i] = address[i];
	}
#pragma unroll
	for (int i = 10; i < 22; i++) {
		input[i] = block[i - 10];
	}
#pragma unroll
	for (int i = 22; i < 24; i++) {
		input[i] = prefix[i-22];
	}
#pragma unroll
	for (int i = 24; i < 34; i++) {
		input[i] = ((nonce >> ((i - 24) * 5)) & 31) + 48;
	}
	digest(input, 34, hashed);
	long score = hashToLong(hashed);
	if (score < work) {
#pragma unroll
		for (int i = 0; i < 10; i++) {
			output[i] = address[i];
		}
#pragma unroll
		for (int i = 10; i < 22; i++) {
			output[i] = block[i - 10];
		}
#pragma unroll
		for (int i = 22; i < 24; i++) {
			output[i] = prefix[i-22];
		}
#pragma unroll
		for (int i = 24; i < 34; i++) {
			output[i] = ((nonce >> ((i - 24) * 5)) & 31) + 48;
		}
	}
}
` + "\x00"

type deviceInfo struct {
	deviceId   cl.DeviceId
	name       string
	vendor     string
	deviceType string
}

func getDevicesList() []deviceInfo {
	var deviceIds = make([]cl.DeviceId, 100)
	var numDevices uint32

	err := cl.GetDeviceIDs(nil, cl.DEVICE_TYPE_ALL, uint32(len(deviceIds)),
		&deviceIds[0], &numDevices)
	if err != cl.SUCCESS {
		log.Fatal("Failed to create device group")
	}

	var allDevices = make([]deviceInfo, numDevices)

	for i := uint32(0); i < numDevices; i++ {
		allDevices[i].deviceId = deviceIds[i]

		var responseLength uint64
		var responseData = make([]byte, BufferSize)

		cl.GetDeviceInfo(deviceIds[i], cl.DEVICE_NAME, BufferSize,
			unsafe.Pointer(&responseData[0]), &responseLength)
		allDevices[i].name = string(responseData[:responseLength])

		cl.GetDeviceInfo(deviceIds[i], cl.DEVICE_VENDOR, BufferSize,
			unsafe.Pointer(&responseData[0]), &responseLength)
		allDevices[i].vendor = string(responseData[:responseLength])

		cl.GetDeviceInfo(deviceIds[i], cl.DEVICE_TYPE, BufferSize,
			unsafe.Pointer(&responseData[0]), &responseLength)
		allDevices[i].deviceType = string(responseData[:responseLength])
	}

	return allDevices
}

func main() {
	// data := make([]float32, BufferSize)

	devices := getDevicesList()
	for _, device := range devices {
		log.Println(device.vendor)
	}

	var err cl.ErrorCode
	var errptr *cl.ErrorCode

	device := devices[1].deviceId

	context := cl.CreateContext(nil, 1, &device, nil, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create context")
	}
	defer cl.ReleaseContext(context)

	cq := cl.CreateCommandQueue(context, device, 0, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create command queue")
	}
	defer cl.ReleaseCommandQueue(cq)

	srcptr := cl.Str(KernelSource)
	program := cl.CreateProgramWithSource(context, 1, &srcptr, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create program")
	}
	defer cl.ReleaseProgram(program)

	err = cl.BuildProgram(program, 1, &device, nil, nil, nil)
	if err != cl.SUCCESS {
		var length uint64
		buffer := make([]byte, BufferSize)

		log.Println("Error: Failed to build program executable!")
		cl.GetProgramBuildInfo(program, device, cl.PROGRAM_BUILD_LOG, uint64(len(buffer)), unsafe.Pointer(&buffer[0]), &length)
		log.Fatal(string(buffer[:length]))
	}

	kernel := cl.CreateKernel(program, cl.Str("krist_miner_basic"+"\x00"), errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create compute kernel")
	}
	defer cl.ReleaseKernel(kernel)

	addressBuf := cl.CreateBuffer(context, cl.MEM_READ_ONLY, BufferSize, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create addressBuf")
	}
	defer cl.ReleaseMemObject(addressBuf)

	blockBuf := cl.CreateBuffer(context, cl.MEM_READ_ONLY, BufferSize, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create blockBuf")
	}
	defer cl.ReleaseMemObject(blockBuf)

	prefixBuf := cl.CreateBuffer(context, cl.MEM_READ_ONLY, BufferSize, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create prefixBuf")
	}
	defer cl.ReleaseMemObject(prefixBuf)

	outputBuf := cl.CreateBuffer(context, cl.MEM_WRITE_ONLY, BufferSize, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create outputBuf")
	}
	defer cl.ReleaseMemObject(outputBuf)

	// __kernel void krist_miner_basic(
	// 	__global const byte* address,	// 10 chars
	// 	__global const byte* block,	// 12 chars
	// 	__global const byte* prefix,	// 2 chars
	// 	const long base,				// convert to 10 chars
	// 	const long work,
	// 	__global byte* output) {

	// Set kernel args
	// count := uint32(BufferSize)
	err = cl.SetKernelArg(kernel, 0, 8, unsafe.Pointer(&addressBuf))
	if err != cl.SUCCESS {
		log.Fatal("Failed to write kernel arg 0")
	}
	err = cl.SetKernelArg(kernel, 1, 8, unsafe.Pointer(&blockBuf))
	if err != cl.SUCCESS {
		log.Fatal("Failed to write kernel arg 1")
	}
	err = cl.SetKernelArg(kernel, 2, 8, unsafe.Pointer(&prefixBuf))
	if err != cl.SUCCESS {
		log.Fatal("Failed to write kernel arg 2")
	}

	address := []byte("khugepoopy")
	err = cl.EnqueueWriteBuffer(cq, addressBuf, cl.TRUE, 0, 10, unsafe.Pointer(&address[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		log.Fatal("failed to write to addressBuf array")
	}

	prefix := []byte("ab")
	err = cl.EnqueueWriteBuffer(cq, prefixBuf, cl.TRUE, 0, 2, unsafe.Pointer(&prefix[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		log.Fatal("failed to write to prefixBuf array")
	}

	block := []byte("abcdefhadwad")
	err = cl.EnqueueWriteBuffer(cq, blockBuf, cl.TRUE, 0, 12, unsafe.Pointer(&block[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		log.Fatal("failed to write to blockBuf array")
	}

	base := 100000
	err = cl.SetKernelArg(kernel, 3, 4, unsafe.Pointer(&base))
	if err != cl.SUCCESS {
		log.Fatal("Failed to write kernel arg 3")
	}

	work := 1000000000
	err = cl.SetKernelArg(kernel, 4, 4, unsafe.Pointer(&work))
	if err != cl.SUCCESS {
		log.Fatal("Failed to write kernel arg 4")
	}

	err = cl.SetKernelArg(kernel, 5, 8, unsafe.Pointer(&outputBuf))
	if err != cl.SUCCESS {
		log.Fatal("Failed to write kernel arg 4")
	}

	zero := uint64(0)
	err = cl.GetKernelWorkGroupInfo(kernel, device, cl.KERNEL_WORK_GROUP_SIZE, 8, unsafe.Pointer(&zero), nil)
	if err != cl.SUCCESS {
		log.Fatal("Failed to get kernel work group info")
	}

	global := uint64(1 << 20)
	local := uint64(64)
	err = cl.EnqueueNDRangeKernel(cq, kernel, 1, &zero, &global, &local, 0, nil, nil)
	if err != cl.SUCCESS {
		log.Fatal("Failed to execute kernel!")
	}

	cl.Finish(cq)

	results := make([]byte, 34)
	err = cl.EnqueueReadBuffer(cq, outputBuf, cl.TRUE, 0, 34, unsafe.Pointer(&results[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		log.Fatal("Failed to read buffer!")
	}

	fmt.Println(string(results))

	res := sha256.Sum256(results)
	fmt.Println(hex.EncodeToString(res[:]))

	log.Println("OK:", results)
}
