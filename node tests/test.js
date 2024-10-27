const axios = require('axios');
const fs = require('fs');
const path = require('path');

const BASE_URL = 'http://localhost:3000';

// Path to the MP3 file
const filePath = path.resolve(__dirname, 'glimpse-of-us.mp3');

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

// Main function to run the entire upload process sequentially
(async function runUploadProcess() {
    const identifier = await initiateFileUpload(filePath);

    if (identifier) {
        await uploadFileChunk(identifier, filePath);
    }
})();

