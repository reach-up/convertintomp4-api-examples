using System.Net.WebSockets;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json.Nodes;

namespace Cim4Examples;

// 3. Realtime updates over a WebSocket instead of polling.
// wss://api.convertintomp4.com/ws/user/jobs streams every job of the key's account:
// {"type": "started" | "progress" | "completed" | "failed", "jobId", "progress"?, "outputUrl"?, "error"?}.
// For multi-task jobs, also pass "webhook_url" when you create the job: you get one signed
// job.finished / job.failed POST when the whole job settles (see VerifyWebhook below).
public static class RealtimeWebsocketUpdates
{
    public static async Task Run()
    {
        using var ws = new ClientWebSocket();
        ws.Options.SetRequestHeader("X-Api-Key", Api.Key);
        ws.Options.SetRequestHeader("User-Agent", Api.UserAgent);
        await ws.ConnectAsync(new Uri(Api.Url.Replace("http", "ws") + "/ws/user/jobs"), CancellationToken.None);
        await Receive(ws); // {"type": "connected"}

        // Start the conversion only once the socket is open, so no event is missed.
        var bytes = await File.ReadAllBytesAsync(Api.SampleFile);
        var created = (await Api.Call(HttpMethod.Post, "/v1/convert",
            Api.FileForm("file", "sample.png", bytes, new() { ["targetFormat"] = "jpg" })))!;
        var jobId = created["jobId"]!.GetValue<string>();
        Console.WriteLine($"Created conversion {jobId}");

        while (ws.State == WebSocketState.Open)
        {
            var msg = await Receive(ws);
            if (msg?["jobId"]?.GetValue<string>() != jobId) continue; // other jobs of the same account
            var type = msg["type"]!.GetValue<string>();
            Console.WriteLine($"  {type} {(msg["progress"] is null ? "" : msg["progress"] + "%")}");
            if (type == "completed") { Console.WriteLine($"Download: {Api.Url}{msg["outputUrl"]}"); break; }
            if (type == "failed") { Console.WriteLine($"Failed: {msg["error"]}"); break; }
        }
        await ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "done", CancellationToken.None);
    }

    static async Task<JsonNode?> Receive(ClientWebSocket ws)
    {
        var buffer = new byte[8192];
        using var ms = new MemoryStream();
        WebSocketReceiveResult r;
        do
        {
            r = await ws.ReceiveAsync(buffer, CancellationToken.None);
            ms.Write(buffer, 0, r.Count);
        } while (!r.EndOfMessage);
        return JsonNode.Parse(Encoding.UTF8.GetString(ms.ToArray()));
    }

    // Check X-ConvertIntoMP4-Signature on the raw body before you trust a webhook.
    // The secret comes from GET /v1/jobs/webhook-secret.
    public static bool VerifyWebhook(byte[] rawBody, string signatureHeader, string secret)
    {
        var expected = "sha256=" + Convert.ToHexString(HMACSHA256.HashData(Encoding.UTF8.GetBytes(secret), rawBody)).ToLowerInvariant();
        return CryptographicOperations.FixedTimeEquals(Encoding.UTF8.GetBytes(expected), Encoding.UTF8.GetBytes(signatureHeader ?? ""));
    }
}
