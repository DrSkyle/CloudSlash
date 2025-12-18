const crypto = require('crypto');

// Configuration
const PRODUCT_ID = "22411";
const PUBLIC_KEY = "pk_fe7225bdb16e44cd1d9495f1be97c";
const SECRET_KEY = "sk_tM&A2j^Q4qlxX}rIjL:f?Xr%pO$rr";
const DEV_ID = "10501"; 


async function signRequest(method, pathURI) {
    const date = new Date().toUTCString();
    const contentMD5 = ""; 
    const contentType = ""; 
    const stringToSign = `${method}\n${contentMD5}\n${contentType}\n${date}\n${pathURI}`;

    const signature = crypto.createHmac('sha256', SECRET_KEY)
                           .update(stringToSign)
                           .digest('base64');
    
    return {
        "Date": date,
        "Authorization": `FS ${PUBLIC_KEY}:${signature}`
    };
}

async function fetchPlans() {
    const path = `/v1/products/${PRODUCT_ID}/plans.json`;
    const url = `https://api.freemius.com${path}`;
    const token = "802fd67∙∙∙∙∙∙∙∙∙∙∙5c4f"; // User provided this (redacted in prompt history, assuming full string was 802fd...5c4f?? No I don't have full string).
    // Ah, wait. The user prompt had "∙∙∙". I don't have the full token. I cannot use this method.
    // ABORT.

    
    try {
        const response = await fetch(url, { headers });
        if (!response.ok) {
            console.error(`Error ${response.status}:`, await response.text());
            return;
        }
        
        const data = await response.json();
        console.log("PLANS FOUND:");
        data.plans.forEach(p => {
            console.log(`- ID: ${p.id} | Name: ${p.name} | Price: ${p.pricing ? JSON.stringify(p.pricing) : "N/A"}`);
        });
    } catch (e) {
        console.error("Fetch failed:", e);
    }
}

fetchPlans();
