const axios = require('axios');
const fs = require('fs');
const path = require('path');
//
// const BASE_URL = 'https://file-service.manojshivagange.tech';
// const PDF_SERVICE_URL = 'https://pdf-service.manojshivagange.tech';
const BASE_URL = 'http://localhost:3000';
const PDF_SERVICE_URL = 'http://localhost:8080';
// Path to the MP3 file
const filePath = path.resolve(__dirname, '2022IMT070.pdf');
const filePath2 = path.resolve(__dirname, 'glimpse-of-us.mp3');

// Function to initiate the file upload
async function initiateFileUpload(filePath) {
    const fileStat = fs.statSync(filePath);

    const uploadRequest = {
        hash: '12345abcde', // You may calculate a real hash here if needed
        name: path.basename(filePath),
        size: fileStat.size
    };

    try {
        const response = await axios.post(`${BASE_URL}/upload/initiate`, uploadRequest);
        console.log('Initiate file upload response:', response.data);
        return response.data.eTag;  // Returns the eTag identifier
    } catch (error) {
        console.error('Error initiating file upload:', error.response?.data);
        return null;
    }
}

// Function to upload file chunks
async function uploadFileChunk(identifier, filePath) {
    const fileStat = fs.statSync(filePath);
    const CHUNK_SIZE = 1024 * 1024; // 1 MB
    const fileStream = fs.createReadStream(filePath, { highWaterMark: CHUNK_SIZE });

    let currentByte = 0;

    for await (const chunk of fileStream) {
        const start = currentByte;
        const end = Math.min(currentByte + CHUNK_SIZE, fileStat.size) - 1;
        const contentRange = `bytes ${start}-${end}/${fileStat.size}`;

        try {
            const response = await axios.put(
                `${BASE_URL}/upload/${identifier}`,
                chunk,
                {
                    headers: {
                        'Content-Range': contentRange
                    }
                }
            );
            console.log(`Uploaded chunk: ${contentRange}, Response: ${response.status}`);
        } catch (error) {
            console.error(`Error uploading chunk ${contentRange}:`, error.response?.data);
            return;
        }

        currentByte += CHUNK_SIZE;
    }

    console.log('File upload completed.');
}

async function downloadFile(identifier) {
    url = `${BASE_URL}/download/${identifier}`;
    const response = await axios.get(url, { responseType: 'stream' });
    const fileName = response.headers['content-disposition'].split('filename=')[1];
    const writeStream = fs.createWrite
    Stream(path.resolve(__dirname, fileName));
    response.data.pipe(writeStream);
    writeStream.on('finish', () => {
        console.log('File downloaded successfully.');
    });
}

async function testCompression( etags ) {
    resp  = await axios.post(`${IMG_SERVICE_URL}/convert`, {
        etags: etags,
        operation: 'compress',
        params: {
            quality: 50
        }
    })

    console.log(resp.data)

}

async function testResize( etag ) {
    resp  = await axios.post(`${IMG_SERVICE_URL}/convert`, {
        etags: [etag],
        operation: 'resize',
        params: {
            width: 100,
            height: 100
        }
    })

    console.log(resp.data)

}

async function splitPDf( etag ) {
    resp  = await axios.post(`${PDF_SERVICE_URL}/convert-pdf-to-images`, {
        etags: [etag],
    })

    console.log(resp.data)
}

async function watermarkPDF( etag ) {
    const resp = await axios.post(`${PDF_SERVICE_URL}/add-watermark`, {
        etags: [etag],
        watermarkText: "sample test",
        opacity: 50,
        position: "center",
    });

    console.log(resp.data)
}

async function mergePDF( etags ) {
    resp  = await axios.post(`${PDF_SERVICE_URL}/convert-images-to-pdf`, {
        etags: etags
    })

    console.log(resp.data)
}



async function testZip(etags){
    try{
        const resp = await axios.post(`${BASE_URL}/zip`, {
            etags: etags
        });
        console.log(resp.data);

    }catch (error){
        console.error("Error during zip operation:", error);
    }

}

async function testResizeAndWatermark(etags) {
    try {
        const resp = await axios.post(`${IMG_SERVICE_URL}/convert`, {
            etags: etags,
            operation: 'watermark',
            params: {
                watermark_text: "Sample Watermark",
                password: ""
            }
        });

        console.log(resp.data);
    } catch (error) {
        console.error("Error during resize and watermark operation:", error);
    }
}


// Main function to run the entire upload process sequentially
(async function runUploadProcess() {
   // const id = await initiateFileUpload(filePath);
   //  if (!id) {
   //       return;
   //  }
   //
   //  await uploadFileChunk(id, filePath)
   //
   //  const nid = await initiateFileUpload(filePath2);
   //  if (!id) {
   //      return;
   //  }
   //
   //  await uploadFileChunk(nid, filePath2)

    await testZip(['518ad310-ef4c-4aa0-8f25-88e47786e185', 'e2b9771e-0d96-4902-a7a3-f9079d1dc5b6'])

    //await watermarkPDF('1ade2869-8937-4c89-812f-39f4f2af7561')

   //  //testCompression(['ebbbc421-9a1c-4080-9ede-5e5578519112','2f03c8bd-1cdf-412f-8cdb-53ee1995516b','8a634787-575c-4ee6-91bd-01fb782077cb'])
   //  mergePDF([
   //          '414f2304-9faf-49dd-8279-3d27b9b199e3',
   //          'f4ce5139-2423-4b9f-a3bc-11ad74461944',
   //          'a3745103-257e-4d57-bcc0-831aee5ba29c',
   //          'e25fcccc-391e-4b5c-8cb2-210d3bf7ff3f'
   //      ]
    //  )
   // testResizeAndWatermark([id ,"099ae98a-2abd-42dd-945c-ecd3e2d83f0b"])
    //downloadFile(id);
})();

