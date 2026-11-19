<template>
  <svg viewBox="0 0 800 800" class="earth-illustration" aria-hidden="true">
    <defs>
      <radialGradient id="blog-earth-ocean" cx="35%" cy="20%" r="83%">
        <stop offset="0" stop-color="#6aa7bd" />
        <stop offset=".36" stop-color="#225579" />
        <stop offset=".72" stop-color="#0a2944" />
        <stop offset="1" stop-color="#020916" />
      </radialGradient>
      <radialGradient id="blog-earth-shade" cx="27%" cy="19%" r="84%">
        <stop offset=".18" stop-color="#020916" stop-opacity="0" />
        <stop offset=".75" stop-color="#020916" stop-opacity=".35" />
        <stop offset="1" stop-color="#020611" stop-opacity=".98" />
      </radialGradient>
      <linearGradient id="blog-earth-land" x2=".8" y2="1">
        <stop stop-color="#719386" />
        <stop offset=".5" stop-color="#526a5f" />
        <stop offset="1" stop-color="#213e38" />
      </linearGradient>
      <clipPath id="blog-earth-circle">
        <circle cx="400" cy="400" r="399" />
      </clipPath>
      <filter id="blog-earth-cloud-blur">
        <feGaussianBlur stdDeviation="3" />
      </filter>
    </defs>
    <g clip-path="url(#blog-earth-circle)">
      <circle cx="400" cy="400" r="400" fill="url(#blog-earth-ocean)" />
      <g fill="url(#blog-earth-land)" opacity=".9">
        <path
          d="m93 86 62-28 49 9 30 23 31-4 18 20 37 4 7 29-24 19-28-4-22 19-17 32 4 30-27 12-11-15-26 3-18-20-28-6-2-24-27-11-20-28z"
        />
        <path
          d="m222 239 14 17 20 8 11 21 21 6 13 22-5 13-19-13-12-18-18-7-8-17-18-10z"
        />
        <path
          d="m285 313 30-11 24 13 18 32 35 12 16 25-12 21 2 29-24 25-4 37-20 26-3 35-19 32-17 36-11-4 2-37-12-30 6-30-13-29 7-36-17-27-3-39-16-25z"
        />
        <path
          d="m418 79 26-13 29 6 21 18 18 1 16 18 33-4 18 23 28 6 12 25-8 20 25 24-5 20-28-1-23 14-20-10-14-23-25-1-12-18-20 2-18-24-20 4-10-21-25-9-5-17z"
        />
        <path
          d="m468 215 33 2 21 19 25 8 17 29-8 29-19 21-9 36-23 17-3 29-24 19-15-11-9-34-17-24-6-30-23-25 8-31 19-9 13-26z"
        />
        <path
          d="m561 123 34-28 46 10 30 16 34-3 34 27 14 34 33 9 4 29-31 17-9 30-26-1-18 25-20-15-17-32-24-10-12-25-18 4-21-22-22-2z"
        />
        <path
          d="m611 270 18 9 12 29 20 10 4 21-12 24 11 23-11 11-16-24-5-32-18-18-12-28zM660 418l25-16 35 12 17 31-8 26-29 13-30-9-19-27z"
        />
        <path
          d="m294 44 15-25 32-4 27 17-8 32-23 19-25-7z"
          fill="#b6cad0"
          opacity=".7"
        />
      </g>
      <g
        fill="none"
        stroke="#e1edf1"
        stroke-linecap="round"
        opacity=".32"
        filter="url(#blog-earth-cloud-blur)"
      >
        <path
          d="M30 203c78-39 155 31 238 5s114-52 181-11 127 21 171 2"
          stroke-width="18"
        />
        <path
          d="M87 251c54-17 84 12 120 7m157-115c49 22 78-9 115-1m35 172c50-4 71 23 134 9"
          stroke-width="9"
        />
        <path
          d="M120 434c47-46 107-36 144-8s64 41 122 29m92-31c28 7 46 29 42 57m48-349c52 7 61 28 102 34"
          stroke-width="14"
        />
        <path
          d="M151 556c76-23 123 33 188 25s102-13 138 21 105 33 151 8"
          stroke-width="21"
        />
      </g>
      <circle cx="400" cy="400" r="400" fill="url(#blog-earth-shade)" />
    </g>
  </svg>
</template>
<style lang="scss" scoped>
.earth-illustration {
  display: block;
  width: 100%;
  height: 100%;
}
</style>
