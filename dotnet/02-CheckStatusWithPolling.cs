namespace Cim4Examples;

// 2. Check status with polling: GET /v1/jobs/{id} until the job is completed, failed or cancelled.
public static class CheckStatusWithPolling
{
    public static async Task Run()
    {
        var job = (await Api.Call(HttpMethod.Post, "/v1/jobs", new
        {
            tasks = new Dictionary<string, object>
            {
                ["import-logo"] = new { operation = "import/url", url = Api.SampleUrl },
                ["to-webp"] = new { operation = "convert", input = "import-logo", output_format = "webp" },
                ["export"] = new { operation = "export/url", input = "to-webp" },
            },
        }))!;
        var id = job["id"]!.GetValue<string>();
        Console.WriteLine($"Created job {id}");

        var deadline = DateTime.UtcNow.AddMinutes(10);
        var delay = TimeSpan.FromSeconds(1);
        string[] done = ["completed", "failed", "cancelled"];
        while (!done.Contains(job["status"]!.GetValue<string>()))
        {
            if (DateTime.UtcNow > deadline) throw new TimeoutException("gave up after 10 minutes");
            await Task.Delay(delay);
            // back off: most jobs finish in seconds, long videos take minutes
            delay = TimeSpan.FromMilliseconds(Math.Min(delay.TotalMilliseconds * 1.5, 10_000));
            job = (await Api.Call(HttpMethod.Get, $"/v1/jobs/{id}"))!;
            var progress = string.Join(" ", job["tasks"]!.AsObject().Select(t => $"{t.Value!["progress"]}%"));
            Console.WriteLine($"  {job["status"]} {progress}");
        }
        Api.PrintTasks(job);
        if (job["status"]!.GetValue<string>() == "completed")
            Console.WriteLine($"Download: {job["tasks"]!["export"]!["result"]!["url"]}");
    }
}
