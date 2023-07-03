import React, {useState} from 'react';
import axios from 'axios';

const App: React.FC = () => {
    const [file, setFile] = useState<File | null>(null);
    const [previewURL, setPreviewURL] = useState<string | null>(null);
    const [errorMessage, setErrorMessage] = useState<string | null>(null);

    const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        if (event.target.files && event.target.files.length > 0) {
            const selectedFile = event.target.files[0];
            setFile(selectedFile);
            setPreviewURL(URL.createObjectURL(selectedFile));
        }
    };

    const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault();

        if (!file) {
            setErrorMessage('Please select a file.');
            return;
        }

        setErrorMessage(null);

        const formData = new FormData();
        formData.append('file', file);

        try {
            const response = await axios.post('/upload', formData);
            const data = response.data;

            if (data.success) {
                setPreviewURL(data.previewURL);
            } else {
                setErrorMessage(data.message);
            }
        } catch (error) {
            console.error('Error:', error);
            setErrorMessage('An error occurred while uploading the file.');
        }
    };

    return (
        <div className="container mx-auto p-4">
            <h1 className="text-3xl font-bold mb-4">File Upload and Stream Preview</h1>
            <form onSubmit={handleSubmit}>
                <input
                    type="file"
                    accept=".webm, .mp4"
                    onChange={handleFileChange}
                    className="mb-4"
                />
                <button type="submit" className="bg-blue-500 text-white px-4 py-2 rounded">
                    Upload
                </button>
            </form>
            {errorMessage && <p className="text-red-500">{errorMessage}</p>}
            {previewURL && (
                <video controls className="mt-4">
                    <source src={previewURL} type="video/mp4"/>
                </video>
            )}
        </div>
    );
};

export default App;
