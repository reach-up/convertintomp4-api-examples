namespace Cim4Examples;

// 4. Uploading files.
// A) One call: POST /v1/convert with the file as multipart form data.
// B) In a job: an import/upload task returns a form URL; send the file there and the job continues.
public static class UploadingFiles
{
    public static async Task Run()
    {
        var bytes = await File.ReadAllBytesAsync(Api.SampleFile);

        // A) one call
        var conversion = (await Api.Call(HttpMethod.Post, "/v1/convert",
            Api.FileForm("file", "sample.png", bytes, new() { ["targetFormat"] = "webp" })))!;
        Console.WriteLine($"A) conversion {conversion["jobId"]} {conversion["status"]} status at {conversion["statusUrl"]}");

        // B) upload inside a job
        var job = (await Api.Call(HttpMethod.Post, "/v1/jobs", new
        {
            tasks = new Dictionary<string, object>
            {
                ["upload"] = new { operation = "import/upload", filename = "sample.png" },
                ["to-jpg"] = new { operation = "convert", input = "upload", output_format = "jpg" },
                ["export"] = new { operation = "export/url", input = "to-jpg" },
            },
        }))!;
        var form = job["tasks"]!["upload"]!["result"]!["form"]!; // { url, method: "POST", field: "file" }
        var path = new Uri(form["url"]!.GetValue<string>()).AbsolutePath;
        await Api.Call(new HttpMethod(form["method"]!.GetValue<string>()), path,
            Api.FileForm(form["field"]!.GetValue<string>(), "sample.png", bytes));
        Console.WriteLine($"B) uploaded to job {job["id"]}");

        var done = await Api.WaitForJob(job["id"]!.GetValue<string>());
        Api.PrintTasks(done);
        Console.WriteLine($"Download: {done["tasks"]!["export"]!["result"]?["url"]}");
    }
}
