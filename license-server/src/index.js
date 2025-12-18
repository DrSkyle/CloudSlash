import { Freemius } from '@freemius/sdk';

/**
 * CloudSlash License Server (Freemius Proxy via SDK)
 * Uses @freemius/sdk to verify licenses securely.
 */
export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    // Health Check
    if (request.method === "GET" && url.pathname === "/") {
      return new Response("CloudSlash License Server Online", { status: 200 });
    }

    // Verify Endpoint
    if (request.method === "POST" && url.pathname === "/verify") {
      try {
        const { licenseKey } = await request.json();
        if (!licenseKey)
          return new Response("Missing licenseKey", { status: 400 });

        // validate env vars
        if (
          !env.FREEMIUS_SECRET_KEY ||
          !env.PRODUCT_ID ||
          !env.FREEMIUS_PUBLIC_KEY ||
          !env.FREEMIUS_API_KEY
        ) {
          return new Response(
            "Server Misconfiguration: Missing Freemius Keys",
            { status: 500 }
          );
        }

        // Initialize SDK with env vars
        const freemius = new Freemius({
          productId: Number(env.PRODUCT_ID),
          apiKey: env.FREEMIUS_API_KEY,
          secretKey: env.FREEMIUS_SECRET_KEY,
          publicKey: env.FREEMIUS_PUBLIC_KEY,
        });

        let license;
        try {
          // Use 'license' (singular) and 'retrieveMany'.
          // Args: (filterObject, paginationObject)
          // Freemius 'filter' param expects a string like "key=XYZ".
          // We pass it as { filter: "key=..." } so it spreads into ?filter=key=...
          const licenses = await freemius.api.license.retrieveMany(
            { filter: `key=${licenseKey}` }, 
            { count: 1 }
          );
          
          if (!licenses || licenses.length === 0) {
             return new Response(
              JSON.stringify({ valid: false, reason: "License Not Found" }),
              { headers: { "Content-Type": "application/json" } }
            );
          }
          license = licenses[0];
        } catch (sdkError) {
          // If SDK throws formatted error
           return new Response(
            JSON.stringify({
              valid: false,
              reason: `Freemius SDK Error: ${sdkError.message}`,
              details: JSON.stringify(sdkError)
            }),
            { headers: { "Content-Type": "application/json" } }
          );
        }

        // Validate Status
        const isValid = !license.is_cancelled && !license.is_expired;
        const reason = !isValid
          ? license.is_cancelled
            ? "License Cancelled"
            : "License Expired"
          : "";

        return new Response(
          JSON.stringify({
            valid: isValid,
            plan: license.plan_title || "Pro",
            expiry: license.expiration
              ? new Date(license.expiration).toISOString()
              : null,
            reason: reason,
          }),
          { headers: { "Content-Type": "application/json" } }
        );

      } catch (e) {
        return new Response(`Internal Error: ${e.message}\nStack: ${e.stack}`, { status: 500 });
      }
    }

    return new Response("Not Found", { status: 404 });
  },
};
