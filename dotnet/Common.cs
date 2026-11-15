// Shared helpers for the ConvertIntoMP4 API examples. API reference: https://convertintomp4.com/api-docs
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Cim4Examples;

public class ApiException(int status, string code, string message, TimeSpan? retryAfter)
    : Exception($"{status} {code}: {message}")
{
    public int Status { get; } = status;
    public string Code { get; } = code;
    public TimeSpan? RetryAfter { get; } = retryAfter;
}

public static class Api
{
    public static readonly string Url = (Environment.GetEnvironmentVariable("CIM4_API_URL") ?? "https://api.convertintomp4.com").TrimEnd('/');
    public static readonly string? Key = Environment.GetEnvironmentVariable("CIM4_API_KEY");
    // Sandbox mode (X-Sandbox: true): files up to 1 MB, 100 requests a month per key. CIM4_SANDBOX=0 for live mode.
    public static readonly bool Sandbox = Environment.GetEnvironmentVariable("CIM4_SANDBOX") != "0";
    public const string SampleUrl = "https://convertintomp4.com/icon-192.png";
    public static readonly string SampleFile = Path.Combine("..", "fixtures", "sample.png");
    public const string UserAgent = "convertintomp4-api-examples/1.0 (dotnet)";

    static readonly HttpClient Http = new() { Timeout = TimeSpan.FromSeconds(330) };

    public static void AddAuth(HttpRequestMessage req)
    {
        req.Headers.Add("X-Api-Key", Key);
        req.Headers.UserAgent.ParseAdd(UserAgent);
        if (Sandbox) req.Headers.Add("X-Sandbox", "true");
    }

    // body: a JSON object, or a Func<HttpContent> for multipart (built again on every attempt).
    public static async Task<JsonNode?> Call(HttpMethod method, string path, object? body = null)
    {
        for (var attempt = 1; ; attempt++)
        {
            using var req = new HttpRequestMessage(method, Url + path);
            AddAuth(req);
            req.Content = body switch
            {
                null => null,
                Func<HttpContent> form => form(),
                _ => new StringContent(JsonSerializer.Serialize(body), Encoding.UTF8, "application/json"),
            };
            using var res = await Http.SendAsync(req);
            var text = await res.Content.ReadAsStringAsync();
            var retryAfter = res.Headers.RetryAfter?.Delta;
            if ((int)res.StatusCode == 429 && attempt < 4)
            {
                // Rate limited (sandbox: 1 POST a second). Wait and send the same request again.
                await Task.Delay(retryAfter ?? TimeSpan.FromMilliseconds(1500 * attempt));
                continue;
            }
            JsonNode? json = null;
            try { json = text.Length > 0 ? JsonNode.Parse(text) : null; } catch (JsonException) { }
            if (!res.IsSuccessStatusCode)
            {
                // Every /v1/jobs error has one shape: {"success": false, "error": {"code", "message", "details"?}}
                var err = json?["error"];
                throw new ApiException((int)res.StatusCode,
                    err?["code"]?.GetValue<string>() ?? $"HTTP_{(int)res.StatusCode}",
                    err?["message"]?.GetValue<string>() ?? text[..Math.Min(text.Length, 200)], retryAfter);
            }
            return json;
        }
    }

    public static Func<HttpContent> FileForm(string field, string fileName, byte[] bytes, Dictionary<string, string>? fields = null) => () =>
    {
        var form = new MultipartFormDataContent();
        foreach (var (k, v) in fields ?? []) form.Add(new StringContent(v), k);
        var file = new ByteArrayContent(bytes);
        file.Headers.ContentType = new MediaTypeHeaderValue("application/octet-stream");
        form.Add(file, field, fileName);
        return form;
    };

    // Wait for a job to settle. One /wait call holds at most 90 s; a longer job answers 408 TIMEOUT
    // and we simply ask again.
    public static async Task<JsonNode> WaitForJob(string id)
    {
        var deadline = DateTime.UtcNow.AddMinutes(10);
        while (true)
        {
            try { return (await Call(HttpMethod.Get, $"/v1/jobs/{id}/wait?timeout=90"))!; }
            catch (ApiException e) when (e.Status == 408 && DateTime.UtcNow < deadline) { }
        }
    }

    public static void PrintTasks(JsonNode job)
    {
        foreach (var (name, t) in job["tasks"]!.AsObject())
        {
            var error = t!["error"]?.GetValue<string>();
            Console.WriteLine($"  {name,-12} {t["operation"],-16} {t["status"]}{(error is null ? "" : "  " + error)}");
        }
    }
}
