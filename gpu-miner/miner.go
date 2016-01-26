package main

import (
	"fmt"
	"github.com/go-gl/cl/v1.2/cl"
	"log"
	"math"
	"unsafe"
)

const (
	// Buffer size for communication to devices
	BufferSize = 1024
)

// KernelSource is the source code of the program we're going to run.
var KernelSource = `
#ifndef uint32_t
#define uint32_t unsigned int
#endif

#define H0 0x6a09e667
#define H1 0xbb67ae85
#define H2 0x3c6ef372
#define H3 0xa54ff53a
#define H4 0x510e527f
#define H5 0x9b05688c
#define H6 0x1f83d9ab
#define H7 0x5be0cd19
#define NONCE_LENGTH 11


uint rotr(uint x, int n) {
    if (n < 32) return (x >> n) | (x << (32 - n));
    return x;
}

uint ch(uint x, uint y, uint z) {
    return (x & y) ^ (~x & z);
}

uint maj(uint x, uint y, uint z) {
    return (x & y) ^ (x & z) ^ (y & z);
}

uint sigma0(uint x) {
    return rotr(x, 2) ^ rotr(x, 13) ^ rotr(x, 22);
}

uint sigma1(uint x) {
    return rotr(x, 6) ^ rotr(x, 11) ^ rotr(x, 25);
}

uint gamma0(uint x) {
    return rotr(x, 7) ^ rotr(x, 18) ^ (x >> 3);
}

uint gamma1(uint x) {
    return rotr(x, 17) ^ rotr(x, 19) ^ (x >> 10);
}

void incrementNonce(uint *startNonce) {
	for (uint place = NONCE_LENGTH - 1; place >= 0; place--) {
		if (startNonce[place] < 'z') {
			startNonce[place] = startNonce[place] + 1;
			return;
		} else {
			startNonce[palce] = 'a';
		}
	}
}

__kernel void perform_work_100000(__global uint *header, __global uint *startNonce,
	__global uint maxWork, __global uint *success) {

}

__kernel void sha256_crypt_kernel(uint *data_info, char *plain_key, uint *digest){
    int t, gid, msg_pad;
    int stop, mmod;
    uint i, ulen, item, total;
    uint W[80], temp, A,B,C,D,E,F,G,H,T1,T2;
    uint num_keys = data_info[1];
	int current_pad;

	uint K[64]={
0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
};

	msg_pad=0;

	ulen = data_info[2];
	total = ulen%64>=56?2:1 + ulen/64;

    digest[0] = H0;
	digest[1] = H1;
	digest[2] = H2;
	digest[3] = H3;
	digest[4] = H4;
	digest[5] = H5;
	digest[6] = H6;
	digest[7] = H7;
	for(item=0; item<total; item++)
	{

		A = digest[0];
		B = digest[1];
		C = digest[2];
		D = digest[3];
		E = digest[4];
		F = digest[5];
		G = digest[6];
		H = digest[7];

#pragma unroll
		for (t = 0; t < 80; t++){
		W[t] = 0x00000000;
		}
		msg_pad=item*64;
		if(ulen > msg_pad)
		{
			current_pad = (ulen-msg_pad)>64?64:(ulen-msg_pad);
		}
		else
		{
			current_pad =-1;
		}

		if(current_pad>0)
		{
			i=current_pad;

			stop =  i/4;
			for (t = 0 ; t < stop ; t++){
				W[t] = ((uchar)  plain_key[msg_pad + t * 4]) << 24;
				W[t] |= ((uchar) plain_key[msg_pad + t * 4 + 1]) << 16;
				W[t] |= ((uchar) plain_key[msg_pad + t * 4 + 2]) << 8;
				W[t] |= (uchar)  plain_key[msg_pad + t * 4 + 3];
			}
			mmod = i % 4;
			if ( mmod == 3){
				W[t] = ((uchar)  plain_key[msg_pad + t * 4]) << 24;
				W[t] |= ((uchar) plain_key[msg_pad + t * 4 + 1]) << 16;
				W[t] |= ((uchar) plain_key[msg_pad + t * 4 + 2]) << 8;
				W[t] |=  ((uchar) 0x80) ;
			} else if (mmod == 2) {
				W[t] = ((uchar)  plain_key[msg_pad + t * 4]) << 24;
				W[t] |= ((uchar) plain_key[msg_pad + t * 4 + 1]) << 16;
				W[t] |=  0x8000 ;
			} else if (mmod == 1) {
				W[t] = ((uchar)  plain_key[msg_pad + t * 4]) << 24;
				W[t] |=  0x800000 ;
			} else /*if (mmod == 0)*/ {
				W[t] =  0x80000000 ;
			}

			if (current_pad<56)
			{
				W[15] =  ulen*8 ;
			}
		}
		else if(current_pad <0)
		{
			if( ulen%64==0)
				W[0]=0x80000000;
			W[15]=ulen*8;
		}

		for (t = 0; t < 64; t++) {
			if (t >= 16)
				W[t] = gamma1(W[t - 2]) + W[t - 7] + gamma0(W[t - 15]) + W[t - 16];
			T1 = H + sigma1(E) + ch(E, F, G) + K[t] + W[t];
			T2 = sigma0(A) + maj(A, B, C);
			H = G; G = F; F = E; E = D + T1; D = C; C = B; B = A; A = T1 + T2;
		}
		digest[0] += A;
		digest[1] += B;
		digest[2] += C;
		digest[3] += D;
		digest[4] += E;
		digest[5] += F;
		digest[6] += G;
		digest[7] += H;
	}


}` + "\x00"

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
	data := make([]float32, BufferSize)

	devices := getDevicesList()
	for _, device := range devices {
		log.Println(device.vendor)
	}

	var err cl.ErrorCode
	var errptr *cl.ErrorCode

	device := devices[0].deviceId

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

	kernel := cl.CreateKernel(program, cl.Str("square"+"\x00"), errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create compute kernel")
	}
	defer cl.ReleaseKernel(kernel)

	input := cl.CreateBuffer(context, cl.MEM_READ_ONLY, 4*BufferSize, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create input buffer")
	}
	defer cl.ReleaseMemObject(input)

	output := cl.CreateBuffer(context, cl.MEM_WRITE_ONLY, 4*BufferSize, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		log.Fatal("couldnt create output buffer")
	}
	defer cl.ReleaseMemObject(output)

	// Write data
	err = cl.EnqueueWriteBuffer(cq, input, cl.TRUE, 0, 4*BufferSize, unsafe.Pointer(&data[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		log.Fatal("Failed to write to source array")
	}

	// Set kernel args
	count := uint32(BufferSize)
	err = cl.SetKernelArg(kernel, 0, 8, unsafe.Pointer(&input))
	if err != cl.SUCCESS {
		log.Fatal("Failed to write kernel arg 0")
	}
	err = cl.SetKernelArg(kernel, 1, 8, unsafe.Pointer(&output))
	if err != cl.SUCCESS {
		log.Fatal("Failed to write kernel arg 1")
	}
	err = cl.SetKernelArg(kernel, 2, 4, unsafe.Pointer(&count))
	if err != cl.SUCCESS {
		log.Fatal("Failed to write kernel arg 2")
	}

	local := uint64(0)
	err = cl.GetKernelWorkGroupInfo(kernel, device, cl.KERNEL_WORK_GROUP_SIZE, 8, unsafe.Pointer(&local), nil)
	if err != cl.SUCCESS {
		log.Fatal("Failed to get kernel work group info")
	}

	global := local
	err = cl.EnqueueNDRangeKernel(cq, kernel, 1, nil, &global, &local, 0, nil, nil)
	if err != cl.SUCCESS {
		log.Fatal("Failed to execute kernel!")
	}

	cl.Finish(cq)

	results := make([]float32, BufferSize)
	err = cl.EnqueueReadBuffer(cq, output, cl.TRUE, 0, 4*1024, unsafe.Pointer(&results[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		log.Fatal("Failed to read buffer!")
	}

	success := 0
	notzero := 0
	for i, x := range data {
		if math.Abs(float64(x*x-results[i])) < 0.5 {
			success++
		}
		if results[i] > 0 {
			notzero++
		}
		log.Printf("I/O: %f\t%f", x, results[i])
	}

	log.Printf("%d/%d success", success, BufferSize)
	log.Printf("values not zero: %d", notzero)
}
