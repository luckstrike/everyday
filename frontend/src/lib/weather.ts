import type { WeatherIcon } from './types.ts';

// Which drawing stands for each kind of weather the backend can report. The
// backend picks the kind from the weather code, so this only maps kind to picture.
const iconNames: Record<WeatherIcon, string> = {
  clear: 'weather',
  'clear-night': 'weather-clear-night',
  'partly-cloudy': 'weather-partly-cloudy',
  'partly-cloudy-night': 'weather-partly-cloudy-night',
  cloudy: 'weather-cloudy',
  fog: 'weather-fog',
  rain: 'weather-rain',
  snow: 'weather-snow',
  storm: 'weather-storm',
};

// A kind this display doesn't know yet (say the backend gains one) shows a plain
// cloud instead of a blank space.
export function weatherIconName(icon: string): string {
  return iconNames[icon as WeatherIcon] ?? 'weather-cloudy';
}
