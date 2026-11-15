namespace Cim4Examples;

// 5. Results and errors.
// Results: an export/url task returns {"url", "file": {"name", "size"}}; the URL is signed and expires.
// Errors: {"success": false, "error": {"code", "message", "details"?}} with the HTTP status.
public static class HandlingResultsAndErrors
{
    // Retry only what can succeed later: rate limits and server errors. Never retry a 4xx input error.
    static async Task<T> WithRetry<T>(Func<Task<T>> fn, int attempts = 4)
    {
        for (var i = 1; ; i++)
        {
            try { return await fn(); }
            catch (ApiException e) when ((e.Status == 429 || e.Status >= 500) && i < attempts)
            {
                var wait = e.RetryAfter ?? TimeSpan.FromSeconds(Math.Pow(2, i));
                Console.WriteLine($"  {e.Code}, retrying in {wait.TotalSeconds} s");
                await Task.Delay(wait);
            }
        }
    }

    public static async Task Run()
    {
        // An input error: the task names an input that does not exist.
        try
        {
            await Api.Call(HttpMethod.Post, "/v1/jobs", new
            {
                tasks = new Dictionary<string, object> { ["to-jpg"] = new { operation = "convert", input = "nothing", output_format = "jpg" } },
            });
        }
        catch (ApiException e) { Console.WriteLine($"Expected error: {e.Message}"); }

        // A job that succeeds: download the result.
        var job = (await WithRetry(() => Api.Call(HttpMethod.Post, "/v1/jobs", new
        {
            tasks = new Dictionary<string, object>
            {
                ["import-logo"] = new { operation = "import/url", url = Api.SampleUrl },
                ["to-png"] = new { operation = "thumbnail", input = "import-logo", output_format = "png", options = new { width = 64, height = 64, fit = "crop" } },
                ["export"] = new { operation = "export/url", input = "to-png" },
            },
        })))!;
        var done = await WithRetry(() => Api.WaitForJob(job["id"]!.GetValue<string>()));
        if (done["status"]!.GetValue<string>() != "completed")
        {
            // A failed task carries its own error; tasks that depended on it are cancelled.
            foreach (var (taskName, t) in done["tasks"]!.AsObject())
                if (t!["error"] is not null) Console.WriteLine($"Task {taskName} failed: {t["error"]}");
            Environment.Exit(1);
        }
        var result = done["tasks"]!["export"]!["result"]!;
        using var http = new HttpClient();
        var data = await http.GetByteArrayAsync(result["url"]!.GetValue<string>());
        var name = result["file"]?["name"]?.GetValue<string>() ?? "result.png";
        await File.WriteAllBytesAsync(name, data);
        Console.WriteLine($"Saved {name} {data.Length} bytes");
    }
}
