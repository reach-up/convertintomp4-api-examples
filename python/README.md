# Python examples

Python 3.9 or newer.

```bash
pip install -r requirements.txt
export CIM4_API_KEY=your_key
python 01-tasks-and-jobs.py
```

Shared helpers are in `common.py`. Use `requests` (or another client that sets its own User-Agent): the bare `urllib` default User-Agent is refused by our edge firewall.
