# Climate Insights KPT Function

The Climate Insights Infrastructure Kptfile lists a custom KPT function that was written in order to make managing and updating the Climate Insight solution using KPT easier.

There are functions like `set-names` that could be used in the Kptfile isntead however this requires a different configPath and therefore would require modifying multiple files. The interface.yaml file structure was created to mimic the `apply-setters` function so that this custom function could also setup replacements as well as the `apply-setters` function.
